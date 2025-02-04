/*
   Copyright 2024 Docker Terraform Provider authors

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package provider

import (
	"fmt"
	"testing"

	"github.com/docker/terraform-provider-docker/internal/envvar"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccOrgSettingImageAccessManagement(t *testing.T) {
	orgName := envvar.GetWithDefault(envvar.AccTestOrganization)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// create
				Config: testAccOrgSettingImageAccessManagement(orgName, true, false, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("docker_org_setting_image_access_management.test", "org_name", orgName),
					resource.TestCheckResourceAttr("docker_org_setting_image_access_management.test", "enabled", "true"),
					resource.TestCheckResourceAttr("docker_org_setting_image_access_management.test", "allow_official_images", "false"),
					resource.TestCheckResourceAttr("docker_org_setting_image_access_management.test", "allow_verified_publishers", "false"),
				),
			},
			{
				// import
				Config:        testAccOrgSettingImageAccessManagement(orgName, true, false, false),
				ImportState:   true,
				ImportStateId: orgName,
				ResourceName:  "docker_org_setting_image_access_management.test",
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("docker_org_setting_image_access_management.test", "org_name", orgName),
					resource.TestCheckResourceAttr("docker_org_setting_image_access_management.test", "enabled", "true"),
					resource.TestCheckResourceAttr("docker_org_setting_image_access_management.test", "allow_official_images", "false"),
					resource.TestCheckResourceAttr("docker_org_setting_image_access_management.test", "allow_verified_publishers", "false"),
				),
			},
			{
				// update setting
				Config: testAccOrgSettingImageAccessManagement(orgName, true, false, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("docker_org_setting_image_access_management.test", "org_name", orgName),
					resource.TestCheckResourceAttr("docker_org_setting_image_access_management.test", "enabled", "true"),
					resource.TestCheckResourceAttr("docker_org_setting_image_access_management.test", "allow_official_images", "false"),
					resource.TestCheckResourceAttr("docker_org_setting_image_access_management.test", "allow_verified_publishers", "true"),
				),
			},
			{
				// disable iam
				Config: testAccOrgSettingImageAccessManagement(orgName, false, false, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("docker_org_setting_image_access_management.test", "org_name", orgName),
					resource.TestCheckResourceAttr("docker_org_setting_image_access_management.test", "enabled", "false"),
					resource.TestCheckResourceAttr("docker_org_setting_image_access_management.test", "allow_official_images", "false"),
					resource.TestCheckResourceAttr("docker_org_setting_image_access_management.test", "allow_verified_publishers", "true"),
				),
			},
			{
				// delete
				Config: " ",
			},
		},
	})
}

func testAccOrgSettingImageAccessManagement(orgName string, enabled, allowOfficialImages, allowVerifiedPublishers bool) string {
	return fmt.Sprintf(`
resource "docker_org_setting_image_access_management" "test" {
  org_name                  = "%[1]s"
  enabled                   = %[2]t
  allow_official_images     = %[3]t
  allow_verified_publishers = %[4]t
}
`, orgName, enabled, allowOfficialImages, allowVerifiedPublishers)
}
