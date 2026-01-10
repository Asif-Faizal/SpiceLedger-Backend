package graphqlschema

import (
	"context"
	"errors"
	"time"

	"github.com/Asif-Faizal/SpiceLedger/internal/service"
	"github.com/google/uuid"
	"github.com/graphql-go/graphql"
)

type Engine struct {
	Schema graphql.Schema
}

func NewEngine(
	dashboardSvc *service.DashboardService,
	productSvc *service.ProductService,
	gradeSvc *service.GradeService,
	inventorySvc *service.InventoryService,
	priceSvc *service.PriceService,
) (*Engine, error) {
	dashboardUsers := graphql.NewObject(graphql.ObjectConfig{
		Name: "DashboardUsersSummary",
		Fields: graphql.Fields{
			"Total":            &graphql.Field{Type: graphql.Int},
			"WeeklyNew":        &graphql.Field{Type: graphql.Int},
			"WeeklyChangePct":  &graphql.Field{Type: graphql.Float},
			"MonthlyChangePct": &graphql.Field{Type: graphql.Float},
		},
	})
	dashboardProducts := graphql.NewObject(graphql.ObjectConfig{
		Name: "DashboardProductsSummary",
		Fields: graphql.Fields{
			"Total":            &graphql.Field{Type: graphql.Int},
			"MonthlyChangePct": &graphql.Field{Type: graphql.Float},
		},
	})
	dashboardGrades := graphql.NewObject(graphql.ObjectConfig{
		Name: "DashboardGradesSummary",
		Fields: graphql.Fields{
			"Total": &graphql.Field{Type: graphql.Int},
		},
	})
	dashboardPriceUpdate := graphql.NewObject(graphql.ObjectConfig{
		Name: "DashboardPriceUpdate",
		Fields: graphql.Fields{
			"Date":          &graphql.Field{Type: graphql.String},
			"ProductID":     &graphql.Field{Type: graphql.String},
			"Product":       &graphql.Field{Type: graphql.String},
			"GradeID":       &graphql.Field{Type: graphql.String},
			"Grade":         &graphql.Field{Type: graphql.String},
			"Price":         &graphql.Field{Type: graphql.Float},
			"PreviousDate":  &graphql.Field{Type: graphql.String},
			"PreviousPrice": &graphql.Field{Type: graphql.Float},
			"ChangeDelta":   &graphql.Field{Type: graphql.Float},
			"ChangePercent": &graphql.Field{Type: graphql.Float},
		},
	})
	dashboardResponse := graphql.NewObject(graphql.ObjectConfig{
		Name: "DashboardResponse",
		Fields: graphql.Fields{
			"Date":       &graphql.Field{Type: graphql.String},
			"Users":      &graphql.Field{Type: dashboardUsers},
			"Products":   &graphql.Field{Type: dashboardProducts},
			"Grades":     &graphql.Field{Type: dashboardGrades},
			"TotalItems": &graphql.Field{Type: graphql.Int},
			"PriceUpdates": &graphql.Field{
				Type: graphql.NewList(dashboardPriceUpdate),
			},
		},
	})

	productType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Product",
		Fields: graphql.Fields{
			"ID":          &graphql.Field{Type: graphql.String},
			"Name":        &graphql.Field{Type: graphql.String},
			"Description": &graphql.Field{Type: graphql.String},
			"CreatedAt":   &graphql.Field{Type: graphql.String},
		},
	})
	gradeType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Grade",
		Fields: graphql.Fields{
			"ID":          &graphql.Field{Type: graphql.String},
			"Name":        &graphql.Field{Type: graphql.String},
			"Description": &graphql.Field{Type: graphql.String},
			"ProductID":   &graphql.Field{Type: graphql.String},
			"CreatedAt":   &graphql.Field{Type: graphql.String},
		},
	})
	inventorySnapshotType := graphql.NewObject(graphql.ObjectConfig{
		Name: "InventorySnapshot",
		Fields: graphql.Fields{
			"ProductID":      &graphql.Field{Type: graphql.String},
			"Product":        &graphql.Field{Type: graphql.String},
			"Grade":          &graphql.Field{Type: graphql.String},
			"TotalQuantity":  &graphql.Field{Type: graphql.Float},
			"AverageCost":    &graphql.Field{Type: graphql.Float},
			"TotalCostBasis": &graphql.Field{Type: graphql.Float},
			"MarketPrice":    &graphql.Field{Type: graphql.Float},
			"MarketValue":    &graphql.Field{Type: graphql.Float},
			"UnrealizedPnL":  &graphql.Field{Type: graphql.Float},
			"RealizedPnL":    &graphql.Field{Type: graphql.Float},
		},
	})
	productInventoryType := graphql.NewObject(graphql.ObjectConfig{
		Name: "ProductInventory",
		Fields: graphql.Fields{
			"ProductID":     &graphql.Field{Type: graphql.String},
			"Product":       &graphql.Field{Type: graphql.String},
			"Grades":        &graphql.Field{Type: graphql.NewList(inventorySnapshotType)},
			"TotalQuantity": &graphql.Field{Type: graphql.Float},
			"TotalValue":    &graphql.Field{Type: graphql.Float},
			"TotalCost":     &graphql.Field{Type: graphql.Float},
			"TotalPnL":      &graphql.Field{Type: graphql.Float},
			"TotalPnLPct":   &graphql.Field{Type: graphql.Float},
		},
	})
	overallInventoryType := graphql.NewObject(graphql.ObjectConfig{
		Name: "OverallInventory",
		Fields: graphql.Fields{
			"Snapshots":     &graphql.Field{Type: graphql.NewList(inventorySnapshotType)},
			"Products":      &graphql.Field{Type: graphql.NewList(productInventoryType)},
			"TotalQuantity": &graphql.Field{Type: graphql.Float},
			"TotalCost":     &graphql.Field{Type: graphql.Float},
			"TotalPnL":      &graphql.Field{Type: graphql.Float},
			"TotalPnLPct":   &graphql.Field{Type: graphql.Float},
		},
	})
	dailyPriceType := graphql.NewObject(graphql.ObjectConfig{
		Name: "DailyPrice",
		Fields: graphql.Fields{
			"Date":       &graphql.Field{Type: graphql.String},
			"ProductID":  &graphql.Field{Type: graphql.String},
			"GradeID":    &graphql.Field{Type: graphql.String},
			"PricePerKg": &graphql.Field{Type: graphql.Float},
		},
	})

	rootQuery := graphql.ObjectConfig{Name: "Query", Fields: graphql.Fields{
		"dashboard": &graphql.Field{
			Type: dashboardResponse,
			Args: graphql.FieldConfigArgument{
				"date": &graphql.ArgumentConfig{Type: graphql.String},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				role, _ := p.Context.Value("role").(string)
				if role != "admin" {
					return nil, errors.New("forbidden")
				}
				dateStr, _ := p.Args["date"].(string)
				var dateVal time.Time
				if dateStr == "" {
					dateVal = time.Now()
				} else {
					t, err := time.Parse("2006-01-02", dateStr)
					if err != nil {
						return nil, err
					}
					dateVal = t
				}
				return dashboardSvc.GetDashboard(p.Context, dateVal)
			},
		},
		"products": &graphql.Field{
			Type: graphql.NewList(productType),
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				return productSvc.ListProducts(p.Context)
			},
		},
		"grades": &graphql.Field{
			Type: graphql.NewList(gradeType),
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				return gradeSvc.ListGrades(p.Context)
			},
		},
		"inventoryCurrent": &graphql.Field{
			Type: overallInventoryType,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				userStr, _ := p.Context.Value("user_id").(string)
				uid, err := uuid.Parse(userStr)
				if err != nil {
					return nil, errors.New("unauthorized")
				}
				return inventorySvc.GetCurrentInventory(p.Context, uid)
			},
		},
		"inventoryOnDate": &graphql.Field{
			Type: overallInventoryType,
			Args: graphql.FieldConfigArgument{
				"date": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				userStr, _ := p.Context.Value("user_id").(string)
				uid, err := uuid.Parse(userStr)
				if err != nil {
					return nil, errors.New("unauthorized")
				}
				dateStr := p.Args["date"].(string)
				t, err := time.Parse("2006-01-02", dateStr)
				if err != nil {
					return nil, err
				}
				return inventorySvc.GetInventoryOnDate(p.Context, uid, t)
			},
		},
		"prices": &graphql.Field{
			Type: graphql.NewList(dailyPriceType),
			Args: graphql.FieldConfigArgument{
				"date": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				dateStr := p.Args["date"].(string)
				return priceSvc.GetPricesForDate(p.Context, dateStr)
			},
		},
		"price": &graphql.Field{
			Type: graphql.Float,
			Args: graphql.FieldConfigArgument{
				"date":      &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
				"productId": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
				"gradeId":   &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				dateStr := p.Args["date"].(string)
				pidStr := p.Args["productId"].(string)
				gidStr := p.Args["gradeId"].(string)
				pid, err := uuid.Parse(pidStr)
				if err != nil {
					return nil, err
				}
				gid, err := uuid.Parse(gidStr)
				if err != nil {
					return nil, err
				}
				return priceSvc.GetPrice(p.Context, dateStr, pid, gid)
			},
		},
	}}

	rootMutation := graphql.ObjectConfig{Name: "Mutation", Fields: graphql.Fields{
		"createProduct": &graphql.Field{
			Type: graphql.Boolean,
			Args: graphql.FieldConfigArgument{
				"name":        &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
				"description": &graphql.ArgumentConfig{Type: graphql.String},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				role, _ := p.Context.Value("role").(string)
				if role != "admin" {
					return nil, errors.New("forbidden")
				}
				name := p.Args["name"].(string)
				desc, _ := p.Args["description"].(string)
				err := productSvc.CreateProduct(p.Context, name, desc)
				return err == nil, err
			},
		},
		"createGrade": &graphql.Field{
			Type: graphql.Boolean,
			Args: graphql.FieldConfigArgument{
				"productId":   &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
				"name":        &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
				"description": &graphql.ArgumentConfig{Type: graphql.String},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				role, _ := p.Context.Value("role").(string)
				if role != "admin" {
					return nil, errors.New("forbidden")
				}
				pidStr := p.Args["productId"].(string)
				name := p.Args["name"].(string)
				desc, _ := p.Args["description"].(string)
				pid, err := uuid.Parse(pidStr)
				if err != nil {
					return nil, err
				}
				err = gradeSvc.CreateGrade(p.Context, pid, name, desc)
				return err == nil, err
			},
		},
		"setPrice": &graphql.Field{
			Type: graphql.Boolean,
			Args: graphql.FieldConfigArgument{
				"date":      &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
				"productId": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
				"gradeId":   &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
				"price":     &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.Float)},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				role, _ := p.Context.Value("role").(string)
				if role != "admin" {
					return nil, errors.New("forbidden")
				}
				dateStr := p.Args["date"].(string)
				pidStr := p.Args["productId"].(string)
				gidStr := p.Args["gradeId"].(string)
				price := p.Args["price"].(float64)
				pid, err := uuid.Parse(pidStr)
				if err != nil {
					return nil, err
				}
				gid, err := uuid.Parse(gidStr)
				if err != nil {
					return nil, err
				}
				err = priceSvc.SetPrice(p.Context, dateStr, pid, gid, price)
				return err == nil, err
			},
		},
		"addLot": &graphql.Field{
			Type: graphql.Boolean,
			Args: graphql.FieldConfigArgument{
				"date":       &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
				"productId":  &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
				"gradeId":    &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
				"quantityKg": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.Float)},
				"unitCost":   &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.Float)},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				userStr, _ := p.Context.Value("user_id").(string)
				uid, err := uuid.Parse(userStr)
				if err != nil {
					return nil, errors.New("unauthorized")
				}
				dateStr := p.Args["date"].(string)
				pidStr := p.Args["productId"].(string)
				gidStr := p.Args["gradeId"].(string)
				qty := p.Args["quantityKg"].(float64)
				cost := p.Args["unitCost"].(float64)
				pid, err := uuid.Parse(pidStr)
				if err != nil {
					return nil, err
				}
				gid, err := uuid.Parse(gidStr)
				if err != nil {
					return nil, err
				}
				t, err := time.Parse("2006-01-02", dateStr)
				if err != nil {
					return nil, err
				}
				err = inventorySvc.AddPurchaseLot(p.Context, uid, t, pid, gid, qty, cost)
				return err == nil, err
			},
		},
		"addSale": &graphql.Field{
			Type: graphql.Boolean,
			Args: graphql.FieldConfigArgument{
				"date":       &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
				"productId":  &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
				"gradeId":    &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
				"quantityKg": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.Float)},
				"unitPrice":  &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.Float)},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				userStr, _ := p.Context.Value("user_id").(string)
				uid, err := uuid.Parse(userStr)
				if err != nil {
					return nil, errors.New("unauthorized")
				}
				dateStr := p.Args["date"].(string)
				pidStr := p.Args["productId"].(string)
				gidStr := p.Args["gradeId"].(string)
				qty := p.Args["quantityKg"].(float64)
				price := p.Args["unitPrice"].(float64)
				pid, err := uuid.Parse(pidStr)
				if err != nil {
					return nil, err
				}
				gid, err := uuid.Parse(gidStr)
				if err != nil {
					return nil, err
				}
				t, err := time.Parse("2006-01-02", dateStr)
				if err != nil {
					return nil, err
				}
				err = inventorySvc.AddSale(p.Context, uid, t, pid, gid, qty, price)
				return err == nil, err
			},
		},
	}}

	schema, err := graphql.NewSchema(graphql.SchemaConfig{
		Query:    graphql.NewObject(rootQuery),
		Mutation: graphql.NewObject(rootMutation),
	})
	if err != nil {
		return nil, err
	}
	return &Engine{Schema: schema}, nil
}

func (e *Engine) Exec(ctx context.Context, query string, variables map[string]interface{}) *graphql.Result {
	return graphql.Do(graphql.Params{
		Schema:         e.Schema,
		RequestString:  query,
		VariableValues: variables,
		Context:        ctx,
	})
}
