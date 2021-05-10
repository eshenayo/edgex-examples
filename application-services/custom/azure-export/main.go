package main

import (
	"fmt"
	"os"

	"github.com/edgexfoundry/app-functions-sdk-go/appsdk"
	"github.com/edgexfoundry/app-functions-sdk-go/pkg/transforms"

	azureTransforms "azure-export/pkg/transforms"
)

const (
	serviceKey           = "AzureExport"
	appConfigDeviceNames = "DeviceNames"
)

var counter int

func main() {
	fmt.Println("Starting " + serviceKey + " Application Service...")

	// 1) First thing to do is to create an instance of the EdgeX SDK, giving it a service key
	edgexSdk := &appsdk.AppFunctionsSDK{ServiceKey: serviceKey}

	// 2) Next, we need to initialize the SDK
	if err := edgexSdk.Initialize(); err != nil {
		edgexSdk.LoggingClient.Error(fmt.Sprintf("SDK initialization failed: %v\n", err))
		os.Exit(-1)
	}

	// 3) Shows how to access the application's specific configuration settings.  Since our DeviceNameFilter
	// Function requires the list of device names we would like to search for, we'll go ahead and define that now.
	deviceNames, err := edgexSdk.GetAppSettingStrings(appConfigDeviceNames)
	if err != nil {
		edgexSdk.LoggingClient.Error(err.Error())
		os.Exit(-1)
	}
	edgexSdk.LoggingClient.Debug(fmt.Sprintf("Filtering for devices %v", deviceNames))

	// Load Azure-specific MQTT configuration from App SDK
	// You can also create AzureMQTTConfig struct yourself
	config, err := azureTransforms.LoadAzureMQTTConfig(edgexSdk)

	if err != nil {
		edgexSdk.LoggingClient.Error(fmt.Sprintf("Failed to load Azure MQTT configurations: %v\n", err))
		os.Exit(-1)
	}

	// 4) This is our pipeline configuration, the collection of functions to
	// execute every time an event is triggered.
	edgexSdk.SetFunctionsPipeline(
		transforms.NewFilter(deviceNames).FilterByDeviceName,
		azureTransforms.NewConversion().TransformToAzure,
		azureTransforms.NewAzureMQTTSender(edgexSdk.LoggingClient, config).MQTTSend,
	)

	// 5) Lastly, we'll go ahead and tell the SDK to "start" and begin listening for events
	// to trigger the pipeline.
	err = edgexSdk.MakeItRun()
	if err != nil {
		edgexSdk.LoggingClient.Error("MakeItRun returned error: ", err.Error())
		os.Exit(-1)
	}

	// Do any required cleanup here

	os.Exit(0)
}
