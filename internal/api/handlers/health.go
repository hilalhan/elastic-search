package handlers

import (
	"encoding/json"

	"github.com/gofiber/fiber/v3"
)

// HealthCheck handles GET requests to check the health of the service
// @Summary 	Health Check
// @Description Checks the health of the service and returns a status message
// @Tags 		Health
// @Accept 		json
// @Produce 	json
// @Success 200 {object} map[string]string{}
// @Router 		/health [get]
func Health(c fiber.Ctx) error {
	healthInfo := map[string]string{
		"status": "ok",
	}

	res, err := json.Marshal(healthInfo)
	if err != nil {
		return err
	}
	c.Response().Header.SetContentType("application/json")
	return c.Send(res)
}
