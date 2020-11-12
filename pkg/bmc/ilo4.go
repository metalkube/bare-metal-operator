// Copyright (c) 2016-2018 Hewlett Packard Enterprise Development LP

package bmc

import (
	"net/http"
	"net/url"
)

func init() {
	RegisterFactory("ilo4", newILOAccessDetails, []string{"https"})
}

func newILOAccessDetails(parsedURL *url.URL, disableCertificateVerification bool) (AccessDetails, error) {
	return &iLOAccessDetails{
		bmcType:                        parsedURL.Scheme,
		portNum:                        parsedURL.Port(),
		hostname:                       parsedURL.Hostname(),
		disableCertificateVerification: disableCertificateVerification,
	}, nil
}

type iLOAccessDetails struct {
	bmcType                        string
	portNum                        string
	hostname                       string
	disableCertificateVerification bool
}

func (a *iLOAccessDetails) Type() string {
	return a.bmcType
}

// NeedsMAC returns true when the host is going to need a separate
// port created rather than having it discovered.
func (a *iLOAccessDetails) NeedsMAC() bool {
	// For the inspection to work, we need a MAC address
	// https://github.com/metal3-io/baremetal-operator/pull/284#discussion_r317579040
	return true
}

func (a *iLOAccessDetails) Driver() string {
	return "ilo"
}

func (a *iLOAccessDetails) DisableCertificateVerification() bool {
	return a.disableCertificateVerification
}

// DriverInfo returns a data structure to pass as the DriverInfo
// parameter when creating a node in Ironic. The structure is
// pre-populated with the access information, and the caller is
// expected to add any other information that might be needed (such as
// the kernel and ramdisk locations).
func (a *iLOAccessDetails) DriverInfo(bmcCreds Credentials) map[string]interface{} {

	result := map[string]interface{}{
		"ilo_username": bmcCreds.Username,
		"ilo_password": bmcCreds.Password,
		"ilo_address":  a.hostname,
	}

	if a.disableCertificateVerification {
		result["ilo_verify_ca"] = false
	}

	if a.portNum != "" {
		result["client_port"] = a.portNum
	}

	return result
}

func (a *iLOAccessDetails) BootInterface() string {
	return "ilo-ipxe"
}

func (a *iLOAccessDetails) ManagementInterface() string {
	return ""
}

func (a *iLOAccessDetails) PowerInterface() string {
	return ""
}

func (a *iLOAccessDetails) RAIDInterface() string {
	return "no-raid"
}

func (a *iLOAccessDetails) VendorInterface() string {
	return ""
}

func (a *iLOAccessDetails) SupportsSecureBoot() bool {
	return true
}

func (a *iLOAccessDetails) Validate(bmcCreds Credentials, bmcClient *http.Client) error {
	return nil
}
