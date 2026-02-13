package collector

import (
	"github.com/yusufpapurcu/wmi"
)

// softwareLicensingProduct maps the LicenseStatus field from SoftwareLicensingProduct.
type softwareLicensingProduct struct {
	LicenseStatus uint32
}

// collectLicense queries the Windows activation status via WMI.
func (c *Collector) collectLicense() (string, error) {
	var products []softwareLicensingProduct
	query := `SELECT LicenseStatus FROM SoftwareLicensingProduct WHERE ApplicationID = '55c92734-d682-4d71-983e-d6ec3f16059f' AND PartialProductKey IS NOT NULL`
	if err := wmi.Query(query, &products); err != nil {
		return "Unknown", err
	}

	if len(products) == 0 {
		return "Not Found", nil
	}

	switch products[0].LicenseStatus {
	case 0:
		return "Unlicensed", nil
	case 1:
		return "Licensed", nil
	case 2:
		return "OOBGrace", nil
	case 3:
		return "OOTGrace", nil
	case 4:
		return "NonGenuineGrace", nil
	case 5:
		return "Notification", nil
	case 6:
		return "ExtendedGrace", nil
	default:
		return "Unknown", nil
	}
}
