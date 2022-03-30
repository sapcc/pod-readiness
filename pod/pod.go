package pod

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/labstack/echo"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type (
	Readiness struct {
		sync.Mutex
		Ready  bool `json:"ready"`
		status map[string]bool
	}

	Pod struct {
		Readiness *Readiness `json:"readiness"`
	}
)

func New() *Pod {
	return &Pod{&Readiness{Ready: true, status: make(map[string]bool)}}
}

func (p *Pod) getAPI() *echo.Echo {
	e := echo.New()
	e.GET("/metrics", echo.WrapHandler(promhttp.Handler()))
	e.GET("/healthy", p.GetHealthy())
	gp := e.Group("/pod")
	gp.GET("/readiness", p.GetReadiness())
	gp.PATCH("/readiness", p.PatchReadiness())
	return e
}

func (p *Pod) StartServer() {
	e := p.getAPI()
	e.Logger.Fatal(e.Start(":8080"))
}

func (p *Pod) GetHealthy() echo.HandlerFunc {
	return func(c echo.Context) (err error) {
		if c.Path() == "/healthy" {
			p.Readiness.Lock()
			defer p.Readiness.Unlock()
			if p.Readiness.Ready {
				return c.String(http.StatusOK, "READY")
			}
			return c.String(http.StatusInternalServerError, "NOT READY")
		}
		return
	}
}

func (p *Pod) GetReadiness() echo.HandlerFunc {
	return func(c echo.Context) (err error) {
		if c.Path() == "/pod/readiness" {
			p.Readiness.Lock()
			defer p.Readiness.Unlock()
			return c.JSON(http.StatusOK, p.Readiness)
		}
		return
	}
}

// PatchReadiness will update the pod readiness
// If the query parameter `key` is supplied then it will be used to
// store the readiness value of this key.
// The combination of all readiness values are use to determine overall readiness
func (p *Pod) PatchReadiness() echo.HandlerFunc {
	return func(c echo.Context) (err error) {
		if c.Path() == "/pod/readiness" {
			p.Readiness.Lock()
			defer p.Readiness.Unlock()

			tempReadiness := Readiness{}
			if err := json.NewDecoder(c.Request().Body).Decode(&tempReadiness); err != nil {
				return err
			}
			key := "default"
			if c.QueryParams().Has("key") {
				// set the readiness for a specific key
				key = c.QueryParams().Get("key")
				if key == "default" {
					return c.String(http.StatusBadRequest, "query parameter `key` must not be 'default'")
				}
			}
			p.Readiness.status[key] = tempReadiness.Ready

			// determine overall readiness
			ready := true
			for _, v := range p.Readiness.status {
				if v == false {
					ready = false
				}
			}
			p.Readiness.Ready = ready

			return c.JSON(http.StatusOK, p.Readiness)
		}
		return
	}
}
