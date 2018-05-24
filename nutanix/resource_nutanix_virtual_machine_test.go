package nutanix

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccNutanixVirtualMachine_basic(t *testing.T) {
	r := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckNutanixVirtualMachineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNutanixVMConfig(r),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNutanixVirtualMachineExists("nutanix_virtual_machine.vm1"),
					resource.TestCheckResourceAttr("nutanix_virtual_machine.vm1", "hardware_clock_timezone", "UTC"),
					resource.TestCheckResourceAttr("nutanix_virtual_machine.vm1", "power_state", "ON"),
					resource.TestCheckResourceAttr("nutanix_virtual_machine.vm1", "memory_size_mib", "2048"),
					resource.TestCheckResourceAttr("nutanix_virtual_machine.vm1", "num_sockets", "1"),
					resource.TestCheckResourceAttr("nutanix_virtual_machine.vm1", "num_vcpus_per_socket", "1"),
				),
			},
		},
	})
}

func testAccCheckNutanixVirtualMachineExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		return nil
	}
}

func testAccCheckNutanixVirtualMachineDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "nutanix_virtual_machine" {
			continue
		}
		for {
			_, err := conn.API.V3.GetVM(rs.Primary.ID)
			if err != nil {
				if strings.Contains(fmt.Sprint(err), "ENTITY_NOT_FOUND") {
					return nil
				}
				return err
			}
			time.Sleep(3000 * time.Millisecond)
		}

	}

	return nil

}

func testAccNutanixVMConfig(r int) string {
	return fmt.Sprint(`
resource "nutanix_category_key" "test-category-key"{
    name = "app-suppport-1"
		description = "App Support Category Key"
}


resource "nutanix_category_value" "test"{
    name = "${nutanix_category_key.test-category-key.id}"
	description = "Test Category Value"
	value = "test-value"
}

data "nutanix_clusters" "clusters" {
  metadata = {
    length = 2
  }
}

output "cluster" {
  value = "${data.nutanix_clusters.clusters.entities.0.metadata.uuid}"
}


resource "nutanix_virtual_machine" "vm1" {
  name = "test-dou"

  categories = [{
	  name   = "${nutanix_category_key.test-category-key.id}"
	  value = "${nutanix_category_value.test.id}"
  }]

  cluster_reference = {
	  kind = "cluster"
	  uuid = "${data.nutanix_clusters.clusters.entities.0.metadata.uuid}"
  }

  num_vcpus_per_socket = 1
  num_sockets          = 1
  memory_size_mib      = 2048
  power_state          = "ON"
}
`)
}