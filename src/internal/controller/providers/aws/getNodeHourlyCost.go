package aws

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	ec2 "github.com/aws/aws-sdk-go-v2/service/ec2"
	pricing "github.com/aws/aws-sdk-go-v2/service/pricing"
	pricingTypes "github.com/aws/aws-sdk-go-v2/service/pricing/types"
	. "github.com/vlasov-y/moneypod/internal/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/record"
	"k8s.io/utils/ptr"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

func (provider *Provider) GetNodeHourlyCost(ctx context.Context, r record.EventRecorder, node *corev1.Node) (hourlyCost float64, err error) {
	log := logf.FromContext(ctx)

	// Get instanceID
	var instanceID string
	if instanceID, err = provider.getInstanceID(ctx, r, node); err != nil {
		return
	}

	// Authorize AWS
	var awsConfig aws.Config
	if awsConfig, err = config.LoadDefaultConfig(ctx); err != nil {
		log.V(1).Error(err, "failed to load AWS config")
		return
	}
	clientEc2 := ec2.NewFromConfig(awsConfig)
	clientPricing := pricing.NewFromConfig(awsConfig, func(o *pricing.Options) {
		o.Region = "us-east-1" // Pricing API only available in us-east-1
	})

	// Describe the instance
	var describe *ec2.DescribeInstancesOutput
	if describe, err = clientEc2.DescribeInstances(ctx, &ec2.DescribeInstancesInput{
		InstanceIds: []string{instanceID},
	}); err != nil {
		log.V(1).Error(err, "failed to describe the instance")
		r.Eventf(node, corev1.EventTypeWarning, "DescribeEC2InstanceFailed", err.Error())
		return
	}

	// Get instance price
	for _, reservation := range describe.Reservations {
		for _, instance := range reservation.Instances {
			// If instance is spot - get spot reservation price
			if instance.SpotInstanceRequestId != nil {
				log.V(2).Info("instance has a spot request")
				// Get spot price from spot instance request
				var spotResult *ec2.DescribeSpotInstanceRequestsOutput
				if spotResult, err = clientEc2.DescribeSpotInstanceRequests(ctx, &ec2.DescribeSpotInstanceRequestsInput{
					SpotInstanceRequestIds: []string{*instance.SpotInstanceRequestId},
				}); err != nil {
					log.V(1).Error(err, "failed to describe spot instance request")
					r.Eventf(node, corev1.EventTypeWarning, "DescribeSpotRequestFailed", err.Error())
					return
				}

				if len(spotResult.SpotInstanceRequests) > 0 {
					spotPrice := spotResult.SpotInstanceRequests[0].SpotPrice
					log.V(2).Info("spot price is defined", "price", spotPrice)
					// Convert hourly cost to float
					if hourlyCost, err = strconv.ParseFloat(*spotPrice, 64); err != nil {
						msg := fmt.Sprintf("failed to parse the spot price: %s", *spotPrice)
						log.V(1).Error(err, msg)
						return
					}
					log.V(1).Info(fmt.Sprintf("spot instance price: %s", *spotPrice))
					r.Eventf(node, corev1.EventTypeNormal, "HourlyCost", *spotPrice)
				} else {
					log.V(2).Info("queried spot requests without an error, but no spot requests are in the list")
					// Spot instance request may not appear instantly, we will try again later
					return hourlyCost, ErrRequestRequeue
				}
			} else {
				log.V(2).Info("instance has no spot request, treating as an on-demand")
				// If instance is on-demand - get the price for instance type in the region
				var priceResult *pricing.GetProductsOutput
				region := (*instance.Placement.AvailabilityZone)[:len(*instance.Placement.AvailabilityZone)-1]
				log.V(2).Info("instance region", "region", region)
				pricingInput := &pricing.GetProductsInput{
					ServiceCode: ptr.To("AmazonEC2"),
					Filters: []pricingTypes.Filter{
						{
							Field: ptr.To("instanceType"),
							Value: ptr.To(string(instance.InstanceType)),
							Type:  pricingTypes.FilterTypeTermMatch,
						},
						{
							Field: ptr.To("regionCode"),
							Value: ptr.To(region),
							Type:  pricingTypes.FilterTypeTermMatch,
						},
						{
							Field: ptr.To("operatingSystem"),
							Value: ptr.To(func() string {
								if instance.Platform == "windows" {
									return "Windows"
								}
								return "Linux"
							}()),
							Type: pricingTypes.FilterTypeTermMatch,
						},
						{
							Field: ptr.To("capacitystatus"),
							Value: ptr.To("Used"),
							Type:  pricingTypes.FilterTypeTermMatch,
						},
						{
							Field: ptr.To("preInstalledSw"),
							Value: ptr.To("NA"),
							Type:  pricingTypes.FilterTypeTermMatch,
						},
					},
				}
				// Verbose filters log
				var labels []any
				for _, filter := range pricingInput.Filters {
					labels = append(labels, *filter.Field, *filter.Value)
				}
				log.V(2).Info("pricing request input filters", labels...)

				// Querying pricing API
				if priceResult, err = clientPricing.GetProducts(ctx, pricingInput); err != nil {
					log.V(1).Error(err, "failed to get instance pricing")
					return
				}

				if len(priceResult.PriceList) > 0 {
					var priceData map[string]interface{}
					if err = json.Unmarshal([]byte(priceResult.PriceList[0]), &priceData); err != nil {
						log.V(1).Error(err, "failed to parse pricing JSON")
						return
					}

					terms := priceData["terms"].(map[string]interface{})
					onDemand := terms["OnDemand"].(map[string]interface{})
					for _, term := range onDemand {
						termData := term.(map[string]interface{})
						priceDimensions := termData["priceDimensions"].(map[string]interface{})
						for _, dimension := range priceDimensions {
							dimensionData := dimension.(map[string]interface{})
							pricePerUnit := dimensionData["pricePerUnit"].(map[string]interface{})
							priceStr := pricePerUnit["USD"].(string)

							if hourlyCost, err = strconv.ParseFloat(priceStr, 64); err != nil {
								msg := fmt.Sprintf("failed to parse the on-demand price: %s", priceStr)
								log.V(1).Error(err, msg)
								return
							}
							log.V(1).Info(fmt.Sprintf("on-demand instance price: %s", priceStr))
							r.Eventf(node, corev1.EventTypeNormal, "HourlyCost", priceStr)
							break
						}
						break
					}
				} else {
					msg := "no pricing data found"
					log.V(1).Info(msg, "instanceType", string(instance.InstanceType))
					return hourlyCost, ErrRequestRequeue
				}
			}
		}
	}

	return
}
