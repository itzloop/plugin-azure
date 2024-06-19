package plugin

import (
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	azfake "github.com/Azure/azure-sdk-for-go/sdk/azcore/fake"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v5"
	computefake "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v5/fake"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armsubscriptions"
	subfake "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armsubscriptions/fake"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAZSDK(t *testing.T) {
	t.Run("subscriptions", func(t *testing.T) {
		azsdk := createSDK(t)
		assert.Len(t, azsdk.subIDs, 1)
		assert.EqualValues(t, "fake-sub", *azsdk.subIDs[0].SubscriptionID)
		assert.EqualValues(t, "fake-tenant", *azsdk.subIDs[0].TenantID)
	})

	t.Run("vms", func(t *testing.T) {
		azsdk := createSDK(t)
		resp, err := azsdk.ListVms()
		require.NoError(t, err)

        vms, ok := resp["fake-sub"]
        assert.True(t, ok, "fake-sub must exist")
        assert.Len(t, vms, 1)
        assert.EqualValues(t, "fake-vm", *vms[0].ID)
	})
}

func createSDK(t *testing.T) *AzureSDKWrapper {

	sdk, err := NewAzureSDKWrapper(AzureSDKParams{
		TokenCredtials: &azfake.TokenCredential{},
		SubscriptionClientOptions: &azcore.ClientOptions{
			Transport: fakeSubTransport(),
		},
		ComputeClientOptions: &azcore.ClientOptions{
			Transport: fakeVMTransport(),
		},
	})
	require.NoError(t, err)

	return sdk
}

func fakeVMTransport() *computefake.VirtualMachinesServerTransport {
	return computefake.NewVirtualMachinesServerTransport(&computefake.VirtualMachinesServer{
		NewListAllPager: func(options *armcompute.VirtualMachinesClientListAllOptions) (resp azfake.PagerResponder[armcompute.VirtualMachinesClientListAllResponse]) {
			resp.AddPage(200, armcompute.VirtualMachinesClientListAllResponse{
				VirtualMachineListResult: armcompute.VirtualMachineListResult{
					Value: []*armcompute.VirtualMachine{
						{ID: newPtr("fake-vm")},
					},
				},
			}, nil)
			return
		},
	})
}

func fakeSubTransport() *subfake.ServerTransport {
	return subfake.NewServerTransport(&subfake.Server{
		NewListPager: func(options *armsubscriptions.ClientListOptions) (resp azfake.PagerResponder[armsubscriptions.ClientListResponse]) {
			resp.AddPage(200, armsubscriptions.ClientListResponse{
				SubscriptionListResult: armsubscriptions.SubscriptionListResult{
					NextLink: nil,
					Value: []*armsubscriptions.Subscription{
						{
							SubscriptionID: newPtr("fake-sub"),
							TenantID:       newPtr("fake-tenant"),
						},
					},
				},
			}, nil)
			return
		},
	})
}

func newPtr[T any](t T) *T {
	return &t
}
