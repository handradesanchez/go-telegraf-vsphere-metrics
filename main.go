package main

import (
    "fmt"
    "log"
    "os"
    "path/filepath"
    "time"

    "github.com/influxdata/telegraf/config"
    "github.com/influxdata/telegraf/plugins/inputs/vsphere"
    "github.com/influxdata/telegraf/testutil"
)

func main() {
    // Read username, password, and vCenter server from environment variables
    username := os.Getenv("QA_VCENTER_USERNAME")
    password := os.Getenv("QA_VCENTER_PASSWORD")
    vcenter := os.Getenv("VCSA_SERVER")

    // Ensure the environment variables are set
    if username == "" || password == "" || vcenter == "" {
        log.Fatal("QA_VCENTER_USERNAME, QA_VCENTER_PASSWORD, and VCSA_SERVER environment variables must be set")
    }

    // Create a vsphere plugin instance
    vspherePlugin := &vsphere.VSphere{
        Vcenters:             []string{"https://" + vcenter + "/sdk"},
        Username:             config.NewSecret([]byte(username)),
        Password:             config.NewSecret([]byte(password)),
        ForceDiscoverOnInit:  true,
        HostMetricExclude:    []string{"*"},
        ClusterMetricExclude: []string{"*"},
        DatacenterMetricExclude: []string{"*"},
        DatastoreMetricExclude:  []string{"*"},
        CollectConcurrency:   8,
        DiscoverConcurrency:  4,
        MaxQueryMetrics:      -1,
        VMMetricInclude: []string{
            "cpu.usagemhz.average",
            "cpu.usage.average",
        },
        VMInstances: false,
        Timeout:     config.Duration(60 * time.Second),
    }

    // Create a test accumulator
    acc := &testutil.Accumulator{}

    // Gather metrics
    err := vspherePlugin.Gather(acc)
    if err != nil {
        log.Fatalf("Error gathering metrics: %v", err)
    }

    // Log all gathered metrics
    for _, metric := range acc.Metrics {
        log.Printf("Metric Name: %s, Fields: %v, Tags: %v", metric.Measurement, metric.Fields, metric.Tags)
    }

    // Retrieve the specific metric
    var metricValue float64
    for _, metric := range acc.Metrics {
        if metric.Measurement == "cpu.usagemhz.average" {
            if value, ok := metric.Fields["value"].(float64); ok {
                metricValue = value
            } else {
                log.Printf("Metric 'cpu.usagemhz.average' found but 'value' field is not a float64")
            }
            break
        }
    }

    // Get the current working directory
    currentDir, err := os.Getwd()
    if err != nil {
        log.Fatalf("Error getting current directory: %v", err)
    }

    // Construct the file path
    filePath := filepath.Join(currentDir, "metrics.txt")

    // Write the metric value to the specified file
    err = os.WriteFile(filePath, []byte(fmt.Sprintf("%f", metricValue)), 0644)
    if err != nil {
        log.Fatalf("Error writing to file: %v", err)
    }

    log.Printf("Metric value written to file: %f", metricValue)
}
