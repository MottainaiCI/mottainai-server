# `pcidb` - the Golang PCI DB library [![Build Status](https://travis-ci.org/jaypipes/pcidb.svg?branch=master)](https://travis-ci.org/jaypipes/pcidb)

`pcidb` is a small Golang library for programmatic querying of PCI vendor,
product and class information

## Usage

`pcidb` contains a PCI database inspection and querying facility that allows
developers to query for information about hardware device classes, vendor and
product information.

The `pcidb.New()` function returns a `pcidb.PCIDB` struct. The `pcidb.PCIDB`
struct contains a number of fields that may be queried for PCI information:

* `pcidb.PCIDB.Classes` is a map, keyed by the PCI class ID (a hex-encoded
  string) of pointers to `pcidb.PCIClass` structs, one for each class of PCI
  device known to `pcidb`
* `pcidb.PCIDB.Vendors` is a map, keyed by the PCI vendor ID (a hex-encoded
  string) of pointers to `pcidb.PCIVendor` structs, one for each PCI vendor
  known to `pcidb`
* `pcidb.PCIDB.Products` is a map, keyed by the PCI product ID* (a hex-encoded
  string) of pointers to `pcidb.PCIProduct` structs, one for each PCI product
  known to `pcidb`

**NOTE**: PCI products are often referred to by their "device ID". We use
the term "product ID" in `pcidb` because it more accurately reflects what the
identifier is for: a specific product line produced by the vendor.

### PCI device classes

Let's take a look at the PCI device class information and how to query the PCI
database for class, subclass, and programming interface information.

Each `pcidb.PCIClass` struct contains the following fields:

* `pcidb.PCIClass.Id` is the hex-encoded string identifier for the device
  class
* `pcidb.PCIClass.Name` is the common name/description of the class
* `pcidb.PCIClass.Subclasses` is an array of pointers to
  `pcidb.PCISubclass` structs, one for each subclass in the device class

Each `pcidb.PCISubclass` struct contains the following fields:

* `pcidb.PCISubclass.Id` is the hex-encoded string identifier for the device
  subclass
* `pcidb.PCISubclass.Name` is the common name/description of the subclass
* `pcidb.PCISubclass.ProgrammingInterfaces` is an array of pointers to
  `pcidb.PCIProgrammingInterface` structs, one for each programming interface
   for the device subclass

Each `pcidb.PCIProgrammingInterface` struct contains the following fields:

* `pcidb.PCIProgrammingInterface.Id` is the hex-encoded string identifier for
  the programming interface
* `pcidb.PCIProgrammingInterface.Name` is the common name/description for the
  programming interface

```go
package main

import (
	"fmt"

	"github.com/jaypipes/pcidb"
)

func main() {
	pci, err := pcidb.New()
	if err != nil {
		fmt.Printf("Error getting PCI info: %v", err)
	}

	for _, devClass := range pci.Classes {
		fmt.Printf(" Device class: %v ('%v')\n", devClass.Name, devClass.Id)
        for _, devSubclass := range devClass.Subclasses {
            fmt.Printf("    Device subclass: %v ('%v')\n", devSubclass.Name, devSubclass.Id)
            for _, progIface := range devSubclass.ProgrammingInterfaces {
                fmt.Printf("        Programming interface: %v ('%v')\n", progIface.Name, progIface.Id)
            }
        }
	}
}
```

Example output from my personal workstation, snipped for brevity:

```
...
 Device class: Serial bus controller ('0c')
    Device subclass: FireWire (IEEE 1394) ('00')
        Programming interface: Generic ('00')
        Programming interface: OHCI ('10')
    Device subclass: ACCESS Bus ('01')
    Device subclass: SSA ('02')
    Device subclass: USB controller ('03')
        Programming interface: UHCI ('00')
        Programming interface: OHCI ('10')
        Programming interface: EHCI ('20')
        Programming interface: XHCI ('30')
        Programming interface: Unspecified ('80')
        Programming interface: USB Device ('fe')
    Device subclass: Fibre Channel ('04')
    Device subclass: SMBus ('05')
    Device subclass: InfiniBand ('06')
    Device subclass: IPMI SMIC interface ('07')
    Device subclass: SERCOS interface ('08')
    Device subclass: CANBUS ('09')
...
```

### PCI vendors and products

Let's take a look at the PCI vendor information and how to query the PCI
database for vendor information and the products a vendor supplies.

Each `pcidb.PCIVendor` struct contains the following fields:

* `pcidb.PCIVendor.Id` is the hex-encoded string identifier for the vendor
* `pcidb.PCIVendor.Name` is the common name/description of the vendor
* `pcidb.PCIVendor.Products` is an array of pointers to `pcidb.PCIProduct`
  structs, one for each product supplied by the vendor

Each `pcidb.PCIProduct` struct contains the following fields:

* `pcidb.PCIProduct.VendorId` is the hex-encoded string identifier for the
  product's vendor
* `pcidb.PCIProduct.Id` is the hex-encoded string identifier for the product
* `pcidb.PCIProduct.Name` is the common name/description of the subclass
* `pcidb.PCIProduct.Subsystems` is an array of pointers to
  `pcidb.PCIProduct` structs, one for each "subsystem" (sometimes called
  "sub-device" in PCI literature) for the product

**NOTE**: A subsystem product may have a different vendor than its "parent" PCI
product. This is sometimes referred to as the "sub-vendor".

Here's some example code that demonstrates listing the PCI vendors with the
most known products:

```go
package main

import (
	"fmt"
	"sort"

	"github.com/jaypipes/pcidb"
)

type ByCountProducts []*pcidb.PCIVendor

func (v ByCountProducts) Len() int {
	return len(v)
}

func (v ByCountProducts) Swap(i, j int) {
	v[i], v[j] = v[j], v[i]
}

func (v ByCountProducts) Less(i, j int) bool {
	return len(v[i].Products) > len(v[j].Products)
}

func main() {
	pci, err := pcidb.New()
	if err != nil {
		fmt.Printf("Error getting PCI info: %v", err)
	}

	vendors := make([]*pcidb.PCIVendor, len(pci.Vendors))
	x := 0
	for _, vendor := range pci.Vendors {
		vendors[x] = vendor
		x++
	}

	sort.Sort(ByCountProducts(vendors))

	fmt.Println("Top 5 vendors by product")
	fmt.Println("====================================================")
	for _, vendor := range vendors[0:5] {
		fmt.Printf("%v ('%v') has %d products\n", vendor.Name, vendor.Id, len(vendor.Products))
	}
}
```

which yields (on my local workstation as of July 7th, 2018):

```
Top 5 vendors by product
====================================================
Intel Corporation ('8086') has 3389 products
NVIDIA Corporation ('10de') has 1358 products
Advanced Micro Devices, Inc. [AMD/ATI] ('1002') has 886 products
National Instruments ('1093') has 601 products
Chelsio Communications Inc ('1425') has 525 products
```

The following is an example of querying the PCI product and subsystem
information to find the products which have the most number of subsystems that
have a different vendor than the top-level product. In other words, the two
products which have been re-sold or re-manufactured with the most number of
different companies.

```go
package main

import (
	"fmt"
	"sort"

	"github.com/jaypipes/pcidb"
)

type ByCountSeparateSubvendors []*pcidb.PCIProduct

func (v ByCountSeparateSubvendors) Len() int {
	return len(v)
}

func (v ByCountSeparateSubvendors) Swap(i, j int) {
	v[i], v[j] = v[j], v[i]
}

func (v ByCountSeparateSubvendors) Less(i, j int) bool {
	iVendor := v[i].VendorId
	iSetSubvendors := make(map[string]bool, 0)
	iNumDiffSubvendors := 0
	jVendor := v[j].VendorId
	jSetSubvendors := make(map[string]bool, 0)
	jNumDiffSubvendors := 0

	for _, sub := range v[i].Subsystems {
		if sub.VendorId != iVendor {
			iSetSubvendors[sub.VendorId] = true
		}
	}
	iNumDiffSubvendors = len(iSetSubvendors)

	for _, sub := range v[j].Subsystems {
		if sub.VendorId != jVendor {
			jSetSubvendors[sub.VendorId] = true
		}
	}
	jNumDiffSubvendors = len(jSetSubvendors)

	return iNumDiffSubvendors > jNumDiffSubvendors
}

func main() {
	pci, err := pcidb.New()
	if err != nil {
		fmt.Printf("Error getting PCI info: %v", err)
	}

	products := make([]*pcidb.PCIProduct, len(pci.Products))
	x := 0
	for _, product := range pci.Products {
		products[x] = product
		x++
	}

	sort.Sort(ByCountSeparateSubvendors(products))

	fmt.Println("Top 2 products by # different subvendors")
	fmt.Println("====================================================")
	for _, product := range products[0:2] {
		vendorId := product.VendorId
		vendor := pci.Vendors[vendorId]
		setSubvendors := make(map[string]bool, 0)

		for _, sub := range product.Subsystems {
			if sub.VendorId != vendorId {
				setSubvendors[sub.VendorId] = true
			}
		}
		fmt.Printf("%v ('%v') from %v\n", product.Name, product.Id, vendor.Name)
		fmt.Printf(" -> %d subsystems under the following different vendors:\n", len(setSubvendors))
		for subvendorId, _ := range setSubvendors {
			subvendor, exists := pci.Vendors[subvendorId]
			subvendorName := "Unknown subvendor"
			if exists {
				subvendorName = subvendor.Name
			}
			fmt.Printf("      - %v ('%v')\n", subvendorName, subvendorId)
		}
	}
}
```

which yields (on my local workstation as of July 7th, 2018):

```
Top 2 products by # different subvendors
====================================================
RTL-8100/8101L/8139 PCI Fast Ethernet Adapter ('8139') from Realtek Semiconductor Co., Ltd.
 -> 34 subsystems under the following different vendors:
      - OVISLINK Corp. ('149c')
      - EPoX Computer Co., Ltd. ('1695')
      - Red Hat, Inc ('1af4')
      - Mitac ('1071')
      - Netgear ('1385')
      - Micro-Star International Co., Ltd. [MSI] ('1462')
      - Hangzhou Silan Microelectronics Co., Ltd. ('1904')
      - Compex ('11f6')
      - Edimax Computer Co. ('1432')
      - KYE Systems Corporation ('1489')
      - ZyXEL Communications Corporation ('187e')
      - Acer Incorporated [ALI] ('1025')
      - Matsushita Electric Industrial Co., Ltd. ('10f7')
      - Ruby Tech Corp. ('146c')
      - Belkin ('1799')
      - Allied Telesis ('1259')
      - Unex Technology Corp. ('1429')
      - CIS Technology Inc ('1436')
      - D-Link System Inc ('1186')
      - Ambicom Inc ('1395')
      - AOPEN Inc. ('a0a0')
      - TTTech Computertechnik AG (Wrong ID) ('0357')
      - Gigabyte Technology Co., Ltd ('1458')
      - Packard Bell B.V. ('1631')
      - Billionton Systems Inc ('14cb')
      - Kingston Technologies ('2646')
      - Accton Technology Corporation ('1113')
      - Samsung Electronics Co Ltd ('144d')
      - Biostar Microtech Int'l Corp ('1565')
      - U.S. Robotics ('16ec')
      - KTI ('8e2e')
      - Hewlett-Packard Company ('103c')
      - ASUSTeK Computer Inc. ('1043')
      - Surecom Technology ('10bd')
Bt878 Video Capture ('036e') from Brooktree Corporation
 -> 30 subsystems under the following different vendors:
      - iTuner ('aa00')
      - Nebula Electronics Ltd. ('0071')
      - DViCO Corporation ('18ac')
      - iTuner ('aa05')
      - iTuner ('aa0d')
      - LeadTek Research Inc. ('107d')
      - Avermedia Technologies Inc ('1461')
      - Chaintech Computer Co. Ltd ('270f')
      - iTuner ('aa07')
      - iTuner ('aa0a')
      - Microtune, Inc. ('1851')
      - iTuner ('aa01')
      - iTuner ('aa04')
      - iTuner ('aa06')
      - iTuner ('aa0f')
      - iTuner ('aa02')
      - iTuner ('aa0b')
      - Pinnacle Systems, Inc. (Wrong ID) ('bd11')
      - Rockwell International ('127a')
      - Askey Computer Corp. ('144f')
      - Twinhan Technology Co. Ltd ('1822')
      - Anritsu Corp. ('1852')
      - iTuner ('aa08')
      - Hauppauge computer works Inc. ('0070')
      - Pinnacle Systems Inc. ('11bd')
      - Conexant Systems, Inc. ('14f1')
      - iTuner ('aa09')
      - iTuner ('aa03')
      - iTuner ('aa0c')
      - iTuner ('aa0e')
```

## Developers

Contributions to `pcidb` are welcomed! Fork the repo on GitHub and submit a pull
request with your proposed changes. Or, feel free to log an issue for a feature
request or bug report.

### Running tests

You can run unit tests easily using the `make test` command, like so:


```
[jaypipes@uberbox pcidb]$ make test
go test github.com/jaypipes/pcidb
ok  	github.com/jaypipes/pcidb	0.045s
```
