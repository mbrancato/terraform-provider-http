package http

import (
	"fmt"
	"net/http"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

const testResourceConfig_basic = `
resource "http" "http_test" {
  action {
    create {
      url = "%s/meta_%d.txt"
      method = "%s"
    }
  }
}

output "body" {
  value = "${http.http_test.action.0.create.0.body}"
}
`

func TestResource_http200(t *testing.T) {
	testHttpMock := setUpMockHttpServer()

	defer testHttpMock.server.Close()

	resource.UnitTest(t, resource.TestCase{
		Providers: testProviders,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testResourceConfig_basic, testHttpMock.server.URL, 200, http.MethodGet),
				Check: func(s *terraform.State) error {
					_, ok := s.RootModule().Resources["http.http_test"]
					if !ok {
						return fmt.Errorf("missing resource")
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

const testResourceConfig_update = `
resource "http" "http_test" {
  action {
    create {
      url = "%s/meta_%d.txt"
      method = "%s"
    }

    update {
      url = "%s/meta_%d.txt"
      method = "PUT"
      request_body = jsonencode({"hello":"update"})
    }
  }
}

output "body" {
  value = "${http.http_test.action.0.create.0.body}"
}
`

func TestResource_update(t *testing.T) {
	testHttpMock := setUpMockHttpServer()

	defer testHttpMock.server.Close()

	resource.UnitTest(t, resource.TestCase{
		Providers: testProviders,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testResourceConfig_basic, testHttpMock.server.URL, 200, http.MethodGet),
				Check: func(s *terraform.State) error {
					_, ok := s.RootModule().Resources["http.http_test"]
					if !ok {
						return fmt.Errorf("missing resource")
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
			{
				Config: fmt.Sprintf(testResourceConfig_update, testHttpMock.server.URL, 200, http.MethodGet, testHttpMock.server.URL, 200),
				Check: func(s *terraform.State) error {
					_, ok := s.RootModule().Resources["http.http_test"]
					if !ok {
						return fmt.Errorf("missing resource")
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

const testResourceConfig_delete = `
resource "http" "http_test" {
  action {
    create {
      url = "%s/meta_%d.txt"
      method = "%s"
    }

    delete {
      url = "%s/meta_%d.txt"
      method = "DELETE"
      response_status_code = 204
    }
  }
}

output "body" {
  value = "${http.http_test.action.0.create.0.body}"
}
`

func TestResource_delete(t *testing.T) {
	testHttpMock := setUpMockHttpServer()

	defer testHttpMock.server.Close()

	resource.UnitTest(t, resource.TestCase{
		Providers: testProviders,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testResourceConfig_delete, testHttpMock.server.URL, 200, http.MethodGet, testHttpMock.server.URL, 200),
				Check: func(s *terraform.State) error {
					_, ok := s.RootModule().Resources["http.http_test"]
					if !ok {
						return fmt.Errorf("missing resource")
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

func TestResource_http404(t *testing.T) {
	testHttpMock := setUpMockHttpServer()

	defer testHttpMock.server.Close()

	resource.UnitTest(t, resource.TestCase{
		Providers: testProviders,
		Steps: []resource.TestStep{
			{
				Config:      fmt.Sprintf(testResourceConfig_basic, testHttpMock.server.URL, 404, http.MethodGet),
				ExpectError: regexp.MustCompile("HTTP request error. Response code: 404"),
			},
		},
	})
}

const testResourceConfig_withHeaders = `
resource "http" "http_test" {

  action {
    create {
      url = "%s/restricted/meta_%d.txt"
      request_headers = {
        Authorization = "Zm9vOmJhcg=="
      }
      request_body = jsonencode({"hello":"world"})
    }
  }
}

output "body" {
  value = "${http.http_test.action.0.create.0.body}"
}
`

func TestResource_withHeaders200(t *testing.T) {
	testHttpMock := setUpMockHttpServer()

	defer testHttpMock.server.Close()

	resource.UnitTest(t, resource.TestCase{
		Providers: testProviders,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testResourceConfig_withHeaders, testHttpMock.server.URL, 200),
				Check: func(s *terraform.State) error {
					_, ok := s.RootModule().Resources["http.http_test"]
					if !ok {
						return fmt.Errorf("missing resource")
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

const testResourceConfig_utf8 = `
resource "http" "http_test" {
  action {
    create {
      url = "%s/utf-8/meta_%d.txt"
      request_body = jsonencode({"hello":"world"})
    }
  }
}

output "body" {
  value = "${http.http_test.action.0.create.0.body}"
}
`

func TestResource_utf8(t *testing.T) {
	testHttpMock := setUpMockHttpServer()

	defer testHttpMock.server.Close()

	resource.UnitTest(t, resource.TestCase{
		Providers: testProviders,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testResourceConfig_utf8, testHttpMock.server.URL, 200),
				Check: func(s *terraform.State) error {
					_, ok := s.RootModule().Resources["http.http_test"]
					if !ok {
						return fmt.Errorf("missing resource")
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

const testResourceConfig_utf16 = `
resource "http" "http_test" {

  action {
    create {
      url = "%s/utf-16/meta_%d.txt"
      request_body = jsonencode({"hello":"world"})
    }
  }
}

output "body" {
  value = "${http.http_test.action.0.create.0.body}"
}
`

func TestResource_utf16(t *testing.T) {
	testHttpMock := setUpMockHttpServer()

	defer testHttpMock.server.Close()

	resource.UnitTest(t, resource.TestCase{
		Providers: testProviders,
		Steps: []resource.TestStep{
			{
				Config:      fmt.Sprintf(testResourceConfig_utf16, testHttpMock.server.URL, 200),
				ExpectError: regexp.MustCompile("Content-Type is not a text type. Got: application/json; charset=UTF-16"),
			},
		},
	})
}

const testResourceConfig_error = `
resource "http" "http_test" {

}
`

func TestResource_compileError(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		Providers: testProviders,
		Steps: []resource.TestStep{
			{
				Config:      testResourceConfig_error,
				ExpectError: regexp.MustCompile("config is invalid: \"action\": required field is not set"),
			},
		},
	})
}

func TestResource_method(t *testing.T) {

	testConf := `
resource "http" "http_test" {

  action {
    create {
      url = "%s/meta_%d.txt"
      method = "%s"
      request_body = jsonencode({"hello": "world"})
    }
  }
}

output "body" {
  value = "${http.http_test.action.0.create.0.body}"
}
`

	testHttpMock := setUpMockHttpServer()

	defer testHttpMock.server.Close()

	resource.UnitTest(t, resource.TestCase{
		Providers: testProviders,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testConf, testHttpMock.server.URL, 200, http.MethodPost),
			},
		},
	})
}
