package http

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

const testDataSourceConfig_basic = `
data "http" "http_test" {
  url = "%s/meta_%d.txt"
}

output "body" {
  value = "${data.http.http_test.body}"
}
`

func TestDataSource_http200(t *testing.T) {
	testHttpMock := setUpMockHttpServer()

	defer testHttpMock.server.Close()

	resource.UnitTest(t, resource.TestCase{
		Providers: testProviders,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testDataSourceConfig_basic, testHttpMock.server.URL, 200),
				Check: func(s *terraform.State) error {
					_, ok := s.RootModule().Resources["data.http.http_test"]
					if !ok {
						return fmt.Errorf("missing data resource")
					}

					outputs := s.RootModule().Outputs

					if outputs["body"].Value != "1.0.0" {
						return fmt.Errorf(
							`'body' output is %s; want '1.0.0'`,
							outputs["body"].Value,
						)
					}

					return nil
				},
			},
		},
	})
}

func TestDataSource_http404(t *testing.T) {
	testHttpMock := setUpMockHttpServer()

	defer testHttpMock.server.Close()

	resource.UnitTest(t, resource.TestCase{
		Providers: testProviders,
		Steps: []resource.TestStep{
			{
				Config:      fmt.Sprintf(testDataSourceConfig_basic, testHttpMock.server.URL, 404),
				ExpectError: regexp.MustCompile("HTTP request error. Response code: 404"),
			},
		},
	})
}

const testDataSourceConfig_withHeaders = `
data "http" "http_test" {
  url = "%s/restricted/meta_%d.txt"

  request_headers = {
    "Authorization" = "Zm9vOmJhcg=="
  }
}

output "body" {
  value = "${data.http.http_test.body}"
}
`

func TestDataSource_withHeaders200(t *testing.T) {
	testHttpMock := setUpMockHttpServer()

	defer testHttpMock.server.Close()

	resource.UnitTest(t, resource.TestCase{
		Providers: testProviders,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testDataSourceConfig_withHeaders, testHttpMock.server.URL, 200),
				Check: func(s *terraform.State) error {
					_, ok := s.RootModule().Resources["data.http.http_test"]
					if !ok {
						return fmt.Errorf("missing data resource")
					}

					outputs := s.RootModule().Outputs

					if outputs["body"].Value != "1.0.0" {
						return fmt.Errorf(
							`'body' output is %s; want '1.0.0'`,
							outputs["body"].Value,
						)
					}

					return nil
				},
			},
		},
	})
}

const testDataSourceConfig_utf8 = `
data "http" "http_test" {
  url = "%s/utf-8/meta_%d.txt"
}

output "body" {
  value = "${data.http.http_test.body}"
}
`

func TestDataSource_utf8(t *testing.T) {
	testHttpMock := setUpMockHttpServer()

	defer testHttpMock.server.Close()

	resource.UnitTest(t, resource.TestCase{
		Providers: testProviders,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testDataSourceConfig_utf8, testHttpMock.server.URL, 200),
				Check: func(s *terraform.State) error {
					_, ok := s.RootModule().Resources["data.http.http_test"]
					if !ok {
						return fmt.Errorf("missing data resource")
					}

					outputs := s.RootModule().Outputs

					if outputs["body"].Value != "1.0.0" {
						return fmt.Errorf(
							`'body' output is %s; want '1.0.0'`,
							outputs["body"].Value,
						)
					}

					return nil
				},
			},
		},
	})
}

const testDataSourceConfig_utf16 = `
data "http" "http_test" {
  url = "%s/utf-16/meta_%d.txt"
}

output "body" {
  value = "${data.http.http_test.body}"
}
`

func TestDataSource_utf16(t *testing.T) {
	testHttpMock := setUpMockHttpServer()

	defer testHttpMock.server.Close()

	resource.UnitTest(t, resource.TestCase{
		Providers: testProviders,
		Steps: []resource.TestStep{
			{
				Config:      fmt.Sprintf(testDataSourceConfig_utf16, testHttpMock.server.URL, 200),
				ExpectError: regexp.MustCompile("Content-Type is not a text type. Got: application/json; charset=UTF-16"),
			},
		},
	})
}

const testDataSourceConfig_error = `
data "http" "http_test" {

}
`

func TestDataSource_compileError(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		Providers: testProviders,
		Steps: []resource.TestStep{
			{
				Config:      testDataSourceConfig_error,
				ExpectError: regexp.MustCompile("The argument \"url\" is required, but no definition was found."),
			},
		},
	})
}
