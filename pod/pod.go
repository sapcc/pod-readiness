package pod

import (
	"net/http"
	"sync"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/labstack/echo"
)

type (
	Readiness struct {
		sync.Mutex
		Ready bool `json:"ready"`
	}

	Pod struct {
		Readiness *Readiness `json:"readiness"`
	}
)

func New() (*Pod){
	return &Pod{&Readiness{Ready: true}}
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

func (p *Pod) PatchReadiness() echo.HandlerFunc {
	return func(c echo.Context) (err error) {
		if c.Path() == "/pod/readiness" {
			p.Readiness.Lock()
			defer p.Readiness.Unlock()
			if err := c.Bind(p.Readiness); err != nil {
				return err
			}
			return c.JSON(http.StatusOK, p.Readiness)
		}
		return
	}
}