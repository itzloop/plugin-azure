package plugin

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v5"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armsubscriptions"
	kaytuproto "github.com/kaytu-io/kaytu/pkg/plugin/proto/src/golang"
	"github.com/kaytu-io/kaytu/pkg/plugin/sdk"
)

type AzurePlugin struct {
	// azure sdk
	// azure processor
	// kaytu ui grpc stream
	stream kaytuproto.Plugin_RegisterClient
}

func NewAzurePlugin() *AzurePlugin {
	return &AzurePlugin{}
}

func (azp *AzurePlugin) ReEvaluate(evaluate *kaytuproto.ReEvaluate) {

}

// TODO
func (azp *AzurePlugin) GetConfig() kaytuproto.RegisterConfig {
	return kaytuproto.RegisterConfig{
		Name:     "itzloop/plugin-azure",
		Version:  version,
		Provider: "azure",
		Commands: []*kaytuproto.Command{
			{
				Name:               "azure_vms",
				Description:        "Get optimization suggestions for your Azure Virtual Machines",
				Flags:              []*kaytuproto.Flag{},           // TODO
				DefaultPreferences: []*kaytuproto.PreferenceItem{}, // TODO
				LoginRequired:      true,
			},
		},
	}
}

func (azp *AzurePlugin) StartProcess(cmd string, flags map[string]string, kaytuAccessToken string, jobQueue *sdk.JobQueue) error {
	creds, err := azidentity.NewAzureCLICredential(nil)
	if err != nil {
		return err
	}

	var subID string

	sc, err := armsubscriptions.NewClient(creds, nil)
	if err != nil {
		return err
	}

	// TODO(itzloop): Handle multiple subscription ids??
	// Only use the first one rn.

	slp := sc.NewListPager(nil)
outterLoop:
	for slp.More() {
		// TODO(itzloop): Make these cancelable
		res, err := slp.NextPage(context.Background())
		if err != nil {
			return err
		}

		for _, sub := range res.Value {
			subID = *sub.SubscriptionID
			break outterLoop
		}
	}

	// TODO(itzloop): Debug logging?
	fmt.Printf("using %s as subscription id", subID)

	rcf, err := armresources.NewClientFactory(subID, creds, nil)
	if err != nil {
		return err
	}

	rgclp := rcf.NewResourceGroupsClient().NewListPager(nil)

	for rgclp.More() {
		// TODO(itzloop): Make these cancelable
		res, err := rgclp.NextPage(context.Background())
		if err != nil {
			return err
		}

		for _, rg := range res.Value {
			vmc, err := armcompute.NewVirtualMachinesClient(subID, creds, nil)
			if err != nil {
				return err
			}

			// Add an options to specify a list rgs
			// if no option is passed, then use
			// vmc.NewListAllPager()

			vmlp := vmc.NewListPager(*rg.Name, nil)
			for vmlp.More() {
				// vms, err := vmlp.NextPage(context.Background())
				// if err != nil {
				//     return err
				// }

				// for _, vm := range vms.Value {
				//    vm.Properties
				// }
			}

		}
	}

	return nil
}

func (azp *AzurePlugin) SetStream(stream kaytuproto.Plugin_RegisterClient) {
	azp.stream = stream
}
