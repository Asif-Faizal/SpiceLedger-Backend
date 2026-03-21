package graphql

type ProductWithGradesAndPrice struct {
	ID          string            `json:"id" validate:"required,uuid4"`
	Name        string            `json:"name" validate:"required,min=3,max=255"`
	Category    string            `json:"category" validate:"required,oneof=spice others"`
	Description string            `json:"description" validate:"omitempty,min=3,max=255"`
	Status      string            `json:"status" validate:"required,oneof=active inactive"`
	Grades      []*GradeWithPrice `json:"grades,omitempty"`
}

type GradeWithPrice struct {
	ID          string  `json:"id" validate:"required,uuid4"`
	ProductID   string  `json:"product_id" validate:"required,uuid4"`
	Name        string  `json:"name" validate:"required,min=3,max=255"`
	Price       float64 `json:"price" validate:"required"`
	Description string  `json:"description" validate:"omitempty,min=3,max=255"`
	Status      string  `json:"status" validate:"required,oneof=active inactive"`
}
