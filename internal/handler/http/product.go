package http

import (
    "github.com/gofiber/fiber/v2"
    "github.com/Asif-Faizal/SpiceLedger/internal/service"
)

type ProductHandler struct {
    productSvc *service.ProductService
}

func NewProductHandler(productSvc *service.ProductService) *ProductHandler {
    return &ProductHandler{productSvc: productSvc}
}

func (h *ProductHandler) CreateProduct(c *fiber.Ctx) error {
    type Request struct {
        Name        string `json:"name"`
        Description string `json:"description"`
    }
    var req Request
    if err := c.BodyParser(&req); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
    }
    if req.Name == "" {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Product name is required"})
    }
    if err := h.productSvc.CreateProduct(c.Context(), req.Name, req.Description); err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
    }
    return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "Product created successfully"})
}

func (h *ProductHandler) ListProducts(c *fiber.Ctx) error {
    products, err := h.productSvc.ListProducts(c.Context())
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
    }
    return c.JSON(products)
}
