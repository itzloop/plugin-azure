package plugin

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v5"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armsubscriptions"
)

type AzureSDKParams struct {
	TokenCredtials            azcore.TokenCredential
	SubscriptionClientOptions *azcore.ClientOptions
	ComputeClientOptions      *azcore.ClientOptions
}

type AzureSDKWrapper struct {
	creds  azcore.TokenCredential
	subIDs []*armsubscriptions.Subscription
	params AzureSDKParams
}

// NewAzureSDKWrapper TODO
func NewAzureSDKWrapper(params AzureSDKParams) (azsdkWrapper *AzureSDKWrapper, err error) {
	var creds azcore.TokenCredential
	if params.TokenCredtials != nil {
		creds = params.TokenCredtials
	} else {
		// TODO(sina): explore *azidentity.AzureCLICredentialOptions and maybe
		// fill it up
		creds, err = azidentity.NewAzureCLICredential(nil)
	}
	if err != nil {
		return nil, err
	}

	azsdkWrapper = &AzureSDKWrapper{
		creds:  creds,
		params: params,
	}

	// create subscriptions client to get the subid
	var sc *armsubscriptions.Client
	if params.SubscriptionClientOptions == nil {
		// TODO(sina): explore *arm.ClientOptionsL and maybe
		// fill it up
		sc, err = armsubscriptions.NewClient(creds, nil)
	} else {
		sc, err = armsubscriptions.NewClient(creds, &arm.ClientOptions{
			ClientOptions: *params.SubscriptionClientOptions,
		})
	}
	if err != nil {
		return nil, err
	}

	slp := sc.NewListPager(nil)
	for slp.More() {
		// TODO(sina): use context.WithTimeout and receive timeout from params
		slpResp, err := slp.NextPage(context.TODO())
		if err != nil {
			return nil, err
		}

		for _, sub := range slpResp.Value {
			azsdkWrapper.subIDs = append(azsdkWrapper.subIDs, sub)
		}
	}

	// TODO(sina): do we need to get resource groups as well

	return azsdkWrapper, nil

}

func (w *AzureSDKWrapper) ListVms() (map[string][]*armcompute.VirtualMachine, error) {
	var result = make(map[string][]*armcompute.VirtualMachine)
	for _, sub := range w.subIDs {
		vmc, err := armcompute.NewVirtualMachinesClient(
			*sub.SubscriptionID,
			w.creds,
			&arm.ClientOptions{
				ClientOptions: *w.params.ComputeClientOptions,
			},
		)

		if err != nil {
			return nil, err
		}

		// TODO(sina): explore *armcompute.VirtualMachinesClientListAllOptions
		// if we want to use some filters
		listPager := vmc.NewListAllPager(nil)
		for listPager.More() {
			// TODO(sina): use context.WithTimeout and receive timeout from params
			vmResp, err := listPager.NextPage(context.TODO())
			if err != nil {
				return nil, err
			}

			for _, vm := range vmResp.Value {
				_, ok := result[*sub.SubscriptionID]
				if !ok {
					result[*sub.SubscriptionID] = []*armcompute.VirtualMachine{}
				}
                
				result[*sub.SubscriptionID] = append(result[*sub.SubscriptionID], vm)
			}
		}
	}

	return result, nil

}
