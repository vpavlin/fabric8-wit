package controller

import (
	"github.com/almighty/almighty-core/app"
	"github.com/almighty/almighty-core/application"
	"github.com/almighty/almighty-core/errors"
	"github.com/almighty/almighty-core/jsonapi"
	"github.com/almighty/almighty-core/login"
	"github.com/almighty/almighty-core/rest"
	"github.com/almighty/almighty-core/workitem/link"
	"github.com/goadesign/goa"
	//uuid "github.com/satori/go.uuid"
)

// WorkItemLinkCategoryController implements the work-item-link-category resource.
type WorkItemLinkCategoryController struct {
	*goa.Controller
	db application.DB
}

// NewWorkItemLinkCategoryController creates a WorkItemLinkCategoryController.
func NewWorkItemLinkCategoryController(service *goa.Service, db application.DB) *WorkItemLinkCategoryController {
	if db == nil {
		panic("db must not be nil")
	}
	return &WorkItemLinkCategoryController{
		Controller: service.NewController("WorkItemLinkCategoryController"),
		db:         db,
	}
}

// enrichLinkCategorySingle includes related resources in the single's "included" array
func enrichLinkCategorySingle(ctx *workItemLinkContext, single app.WorkItemLinkCategorySingle) error {
	// Add "links" element
	selfURL := rest.AbsoluteURL(ctx.RequestData, ctx.LinkFunc(single.Data.ID))
	single.Data.Links = &app.GenericLinks{
		Self: &selfURL,
	}
	return nil
}

// enrichLinkCategoryList includes related resources in the list's "included" array
func enrichLinkCategoryList(ctx *workItemLinkContext, list *app.WorkItemLinkCategoryList) error {
	// Add "links" element
	for _, data := range list.Data {
		selfURL := rest.AbsoluteURL(ctx.RequestData, ctx.LinkFunc(*data.ID))
		data.Links = &app.GenericLinks{
			Self: &selfURL,
		}
	}
	return nil
}

// Create runs the create action.
func (c *WorkItemLinkCategoryController) Create(ctx *app.CreateWorkItemLinkCategoryContext) error {
	currentUserIdentityID, err := login.ContextIdentity(ctx)
	if err != nil {
		return jsonapi.JSONErrorResponse(ctx, errors.NewUnauthorizedError(err.Error()))
	}
	return application.Transactional(c.db, func(appl application.Application) error {
		modelCategory, err := appl.WorkItemLinkCategories().Create(ctx.Context, ctx.Payload.Data.Attributes.Name, ctx.Payload.Data.Attributes.Description)
		if err != nil {
			jerrors, httpStatusCode := jsonapi.ErrorToJSONAPIErrors(err)
			return ctx.ResponseData.Service.Send(ctx.Context, httpStatusCode, jerrors)
		}
		appCategory := convertLinkCategoryFromModel(*modelCategory)
		linkCtx := newWorkItemLinkContext(ctx.Context, appl, c.db, ctx.RequestData, ctx.ResponseData, app.WorkItemLinkCategoryHref, currentUserIdentityID)
		err = enrichLinkCategorySingle(linkCtx, appCategory)
		if err != nil {
			jerrors, _ := jsonapi.ErrorToJSONAPIErrors(goa.ErrInternal("Failed to enrich link category: %s", err.Error()))
			return ctx.InternalServerError(jerrors)
		}
		ctx.ResponseData.Header().Set("Location", app.WorkItemLinkCategoryHref(appCategory.Data.ID))
		return ctx.Created(&appCategory)
	})
}

// Show runs the show action.
func (c *WorkItemLinkCategoryController) Show(ctx *app.ShowWorkItemLinkCategoryContext) error {
	return application.Transactional(c.db, func(appl application.Application) error {
		modelCategory, err := appl.WorkItemLinkCategories().Load(ctx.Context, ctx.ID)
		if err != nil {
			jerrors, httpStatusCode := jsonapi.ErrorToJSONAPIErrors(err)
			return ctx.ResponseData.Service.Send(ctx.Context, httpStatusCode, jerrors)
		}
		appCategory := convertLinkCategoryFromModel(*modelCategory)
		linkCtx := newWorkItemLinkContext(ctx.Context, appl, c.db, ctx.RequestData, ctx.ResponseData, app.WorkItemLinkCategoryHref, nil)
		err = enrichLinkCategorySingle(linkCtx, appCategory)
		if err != nil {
			jerrors, _ := jsonapi.ErrorToJSONAPIErrors(goa.ErrInternal("Failed to enrich link category: %s", err.Error()))
			return ctx.InternalServerError(jerrors)
		}
		return ctx.OK(&appCategory)
	})
}

// List runs the list action.
func (c *WorkItemLinkCategoryController) List(ctx *app.ListWorkItemLinkCategoryContext) error {
	return application.Transactional(c.db, func(appl application.Application) error {
		modelCategories, err := appl.WorkItemLinkCategories().List(ctx.Context)
		if err != nil {
			jerrors, httpStatusCode := jsonapi.ErrorToJSONAPIErrors(err)
			return ctx.ResponseData.Service.Send(ctx.Context, httpStatusCode, jerrors)
		}
		// convert
		appCategories := app.WorkItemLinkCategoryList{}
		appCategories.Data = make([]*app.WorkItemLinkCategoryData, len(modelCategories))
		for index, value := range modelCategories {
			cat := convertLinkCategoryFromModel(value)
			appCategories.Data[index] = cat.Data
		}
		// TODO: When adding pagination, this must not be len(rows) but
		// the overall total number of elements from all pages.
		appCategories.Meta = &app.WorkItemLinkCategoryListMeta{
			TotalCount: len(modelCategories),
		}
		// Enrich
		linkCtx := newWorkItemLinkContext(ctx.Context, appl, c.db, ctx.RequestData, ctx.ResponseData, app.WorkItemLinkCategoryHref, nil)
		err = enrichLinkCategoryList(linkCtx, &appCategories)
		if err != nil {
			jerrors, _ := jsonapi.ErrorToJSONAPIErrors(goa.ErrInternal("Failed to enrich link categories: %s", err.Error()))
			return ctx.InternalServerError(jerrors)
		}
		return ctx.OK(&appCategories)
	})
}

// Delete runs the delete action.
func (c *WorkItemLinkCategoryController) Delete(ctx *app.DeleteWorkItemLinkCategoryContext) error {
	return application.Transactional(c.db, func(appl application.Application) error {
		err := appl.WorkItemLinkCategories().Delete(ctx.Context, ctx.ID)
		if err != nil {
			jerrors, httpStatusCode := jsonapi.ErrorToJSONAPIErrors(err)
			return ctx.ResponseData.Service.Send(ctx.Context, httpStatusCode, jerrors)
		}
		return ctx.OK([]byte{})
	})
}

// Update runs the update action.
func (c *WorkItemLinkCategoryController) Update(ctx *app.UpdateWorkItemLinkCategoryContext) error {
	currentUserIdentityID, err := login.ContextIdentity(ctx)
	if err != nil {
		return jsonapi.JSONErrorResponse(ctx, errors.NewUnauthorizedError(err.Error()))
	}
	appCategory := app.WorkItemLinkCategorySingle{
		Data: ctx.Payload.Data,
	}
	if appCategory.Data.ID == nil {
		return errors.NewBadParameterError("data.id", appCategory.Data.ID)
	}
	if appCategory.Data.Attributes.Name == nil || *appCategory.Data.Attributes.Name == "" {
		return errors.NewBadParameterError("data.attributes.name", "nil or empty")
	}
	modelCategory := convertLinkCategoryToModel(appCategory)
	return application.Transactional(c.db, func(appl application.Application) error {
		savedModelCategory, err := appl.WorkItemLinkCategories().Save(ctx.Context, modelCategory)
		if err != nil {
			jerrors, httpStatusCode := jsonapi.ErrorToJSONAPIErrors(err)
			return ctx.ResponseData.Service.Send(ctx.Context, httpStatusCode, jerrors)
		}
		// convert to app representation
		savedAppCategory := convertLinkCategoryFromModel(*savedModelCategory)
		// Enrich
		linkCtx := newWorkItemLinkContext(ctx.Context, appl, c.db, ctx.RequestData, ctx.ResponseData, app.WorkItemLinkCategoryHref, currentUserIdentityID)
		err = enrichLinkCategorySingle(linkCtx, savedAppCategory)
		if err != nil {
			jerrors, _ := jsonapi.ErrorToJSONAPIErrors(goa.ErrInternal("Failed to enrich link category: %s", err.Error()))
			return ctx.InternalServerError(jerrors)
		}
		return ctx.OK(&savedAppCategory)
	})
}

// convertLinkCategoryFromModel converts work item link category from model to app representation
func convertLinkCategoryFromModel(t link.WorkItemLinkCategory) app.WorkItemLinkCategorySingle {
	var converted = app.WorkItemLinkCategorySingle{
		Data: &app.WorkItemLinkCategoryData{
			Type: link.EndpointWorkItemLinkCategories,
			ID:   &t.ID,
			Attributes: &app.WorkItemLinkCategoryAttributes{
				Name:        &t.Name,
				Description: t.Description,
				Version:     &t.Version,
			},
		},
	}
	return converted
}

// convertLinkCategoryToModel converts work item link category from app to app representation
func convertLinkCategoryToModel(t app.WorkItemLinkCategorySingle) link.WorkItemLinkCategory {
	var converted = link.WorkItemLinkCategory{}
	if t.Data.ID != nil {
		converted.ID = *t.Data.ID
	}
	if t.Data.Attributes.Version != nil {
		converted.Version = *t.Data.Attributes.Version
	}
	if t.Data.Attributes.Name != nil {
		converted.Name = *t.Data.Attributes.Name
	}
	if t.Data.Attributes.Description != nil {
		converted.Description = t.Data.Attributes.Description
	}
	return converted
}
