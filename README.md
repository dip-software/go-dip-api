[![Maintainability](https://api.codeclimate.com/v1/badges/125caa4282d4d82b84cd/maintainability)](https://codeclimate.com/github/dip-software/go-dip-api/maintainability)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/dip-software/go-dip-api)](https://pkg.go.dev/github.com/dip-software/go-dip-api)

# go-dip-api

A DIP API client library enabling Go programs to interact with various DIP APIs in a simple and uniform way

> [!Important]
> This library is not endorsed, supported or approved by any corporate entity. It is a DIP Software Open Source community managed project. Please do not raise
> SNOW tickets, instead open a issue on the [Github project](https://github.com/dip-software/go-dip-api/issues).

## Supported APIs

The current implement covers only a subset of DIP APIs. Basically, we implement functionality as needed.


- [x] Cartel c.q. Container Host management ([examples](cartel/README.md))
- [x] Clinical Data Repository (CDR)
  - [x] Tenant Onboarding
  - [x] Subscription management
  - [x] FHIR CRUD
  - [x] FHIR Patch
  - [x] STU3
  - [x] R4
- [x] Connect IoT
  - [x] Master Data Management (MDM)
    - [x] Propositions
    - [x] Applications
    - [x] Data Adapter
    - [x] Data Subscribers
    - [x] OAuth2 clients
    - [x] Standard Services
    - [x] Service Actions
    - [x] Service References
    - [x] Storage Classes
    - [x] Device Groups
    - [x] Device Types
    - [x] Regions
    - [x] Buckets
    - [x] Data Types
    - [x] Blob Data Contracts
    - [x] Blob Subscriptions
    - [x] Data Broker Subscriptions
    - [x] Firmware Components
    - [x] Firmware Component Versions
    - [x] OAuth Client Scopes
    - [x] Subscriber Types
    - [x] Resources Limits
    - [x] Authentication Methods
  - [x] Data Broker
    - [ ] Data Items
    - [x] Subscribers
      - [x] SQS
      - [ ] Kinesis
    - [x] Subscriptions
    - [ ] Access Details
  - [x] Blob Repository
    - [x] Blob Metadata
    - [x] Access Policy
    - [x] Access URL
    - [x] Multipart Upload
    - [x] BlobStore Policy management
    - [ ] Topic management
    - [ ] Store Access
    - [ ] Bucket management
    - [ ] Contract management
    - [ ] Subscription management
- [x] Secure Transport Layer (STL) / Edge 
  - [x] Device queries
  - [x] Application Resources management
  - [x] Device configuration management (firewall, logging)
- [x] Public Key Infrastructure (PKI) management
- [x] Identity and Access Management (IAM)
  - [x] Groups
  - [x] Organizations
  - [x] Permissions
  - [x] Roles
  - [x] Role Sharing Policies
  - [x] Users
  - [x] Passwords
  - [x] Propositions
  - [x] Applications
  - [x] Services
  - [x] Devices
  - [x] MFA Policies
  - [x] Password Policies
  - [x] Email Templates
  - [x] SMS Gateways
  - [x] SMS Templates
- [x] Logging ([examples](logging/README.md))
- [x] Auditing ([examples](audit/README.md))
- [x] Telemetry Data Repository (TDR)
  - [x] Contract management
  - [x] Data Item management
- [x] S3Creds Policy management
- [x] DICOM Store
  - [x] Config management
- [x] Notification service
- [x] Service Discovery
- [x] Console settings
  - [ ] Metrics Alerts
  - [x] Metrics Autoscalers
- [x] Docker Registry
  - [x] Service Keys management
  - [x] Namespace management
  - [x] Repository management
- [x] IronIO tasks, codes and schedules management ([examples](iron/README.md))

## Example usage

```go
package main

import (
        "fmt"

        "github.com/dip-software/go-dip-api/iam"
)

func main() {
        client, _ := iam.NewClient(nil, &iam.Config{
                Region:         "us-east",
                Environment:    "client-test",
                OAuth2ClientID: "ClientID",
                OAuth2Secret:   "ClientPWD",
        })
        err := client.Login("iam.login@hospital1.com", "Password!@#")
        if err != nil {
                fmt.Printf("Error logging in: %v\n", err)
                return
        }
        introspect, _, _ := client.Introspect()
        if introspect != nil {
                fmt.Printf("Introspect response: %v\n", introspect)
        }
}
```

## TODO

- Increase API coverage

## Issues

- If you discover an issue: report it on the [issue tracker](https://github.com/dip-software/go-dip-api/issues)

## License

License is MIT. See [LICENSE file](LICENSE.md)
