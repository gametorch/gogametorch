package gametorch

import (
	"context"
	"net/http"
	"net/url"
)

type createLabelBody struct {
	Name  string  `json:"name"`
	Color *string `json:"color,omitempty"`
}

type updateLabelBody struct {
	Name  *string `json:"name,omitempty"`
	Color *string `json:"color,omitempty"`
}

type associateLabelBody struct {
	Name string `json:"name"`
}

type thumbnailBody struct {
	AssetID *string `json:"asset_id"`
}

// ListLabels lists a project's labels.
//
// GET /projects/{project_id}/labels
func (c *Client) ListLabels(ctx context.Context, projectID string) (*LabelsResponse, error) {
	out, err := fetchJSON[LabelsResponse](c, ctx, rateTier2, "GET /projects/{project_id}/labels", concurrencyNone, requestSpec{
		method: http.MethodGet,
		path:   "projects/" + projectID + "/labels",
	})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// CreateLabel creates a label. A project may have at most 10,000 labels. Pass a
// nil color to leave it unset.
//
// POST /projects/{project_id}/labels
func (c *Client) CreateLabel(ctx context.Context, projectID, name string, color *string) (*Label, error) {
	out, err := fetchJSON[Label](c, ctx, rateWrites, "POST /projects/{project_id}/labels", concurrencyNone, requestSpec{
		method: http.MethodPost,
		path:   "projects/" + projectID + "/labels",
		body:   createLabelBody{Name: name, Color: color},
	})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateLabel updates a label's name and/or color.
//
// PATCH /labels/{label_id}
func (c *Client) UpdateLabel(ctx context.Context, labelID string, name, color *string) (*Label, error) {
	out, err := fetchJSON[Label](c, ctx, rateWrites, "PATCH /labels/{label_id}", concurrencyNone, requestSpec{
		method: http.MethodPatch,
		path:   "labels/" + labelID,
		body:   updateLabelBody{Name: name, Color: color},
	})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteLabel deletes a label.
//
// DELETE /labels/{label_id}
func (c *Client) DeleteLabel(ctx context.Context, labelID string) (*OkResponse, error) {
	return fetchAck(c, ctx, rateWrites, "DELETE /labels/{label_id}", concurrencyNone, requestSpec{
		method: http.MethodDelete,
		path:   "labels/" + labelID,
	})
}

// LabelItems lists the sprites, sounds and saved animations tagged with a label.
//
// GET /labels/{label_id}/items
func (c *Client) LabelItems(ctx context.Context, labelID string) (*LabelItems, error) {
	out, err := fetchJSON[LabelItems](c, ctx, rateTier2, "GET /labels/{label_id}/items", concurrencyNone, requestSpec{
		method: http.MethodGet,
		path:   "labels/" + labelID + "/items",
	})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// SetLabelThumbnail sets or clears a label's cover thumbnail. Pass nil to
// clear it.
//
// POST /labels/{label_id}/thumbnail
func (c *Client) SetLabelThumbnail(ctx context.Context, labelID string, assetID *string) (*Label, error) {
	out, err := fetchJSON[Label](c, ctx, rateWrites, "POST /labels/{label_id}/thumbnail", concurrencyNone, requestSpec{
		method: http.MethodPost,
		path:   "labels/" + labelID + "/thumbnail",
		body:   thumbnailBody{AssetID: assetID},
	})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// AssociateAssetLabel adds a label to a sprite asset.
//
// POST /assets/{asset_id}/labels
func (c *Client) AssociateAssetLabel(ctx context.Context, assetID, name string) (*LabelAssociation, error) {
	out, err := fetchJSON[LabelAssociation](c, ctx, rateWrites, "POST /assets/{asset_id}/labels", concurrencyNone, requestSpec{
		method: http.MethodPost,
		path:   "assets/" + assetID + "/labels",
		body:   associateLabelBody{Name: name},
	})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// RemoveAssetLabel removes a label from a sprite asset.
//
// DELETE /assets/{asset_id}/labels?name={name}
func (c *Client) RemoveAssetLabel(ctx context.Context, assetID, name string) (*LabelAssociation, error) {
	out, err := fetchJSON[LabelAssociation](c, ctx, rateWrites, "DELETE /assets/{asset_id}/labels", concurrencyNone, requestSpec{
		method: http.MethodDelete,
		path:   "assets/" + assetID + "/labels",
		query:  url.Values{"name": {name}},
	})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// DismissAssetLabelSuggestion dismisses a suggested label for a sprite asset.
//
// DELETE /assets/{asset_id}/label-suggestions?name={name}
func (c *Client) DismissAssetLabelSuggestion(ctx context.Context, assetID, name string) (*OkResponse, error) {
	return fetchAck(c, ctx, rateWrites, "DELETE /assets/{asset_id}/label-suggestions", concurrencyNone, requestSpec{
		method: http.MethodDelete,
		path:   "assets/" + assetID + "/label-suggestions",
		query:  url.Values{"name": {name}},
	})
}

// AssociateSoundLabel adds a label to a sound asset.
//
// POST /sound-assets/{asset_id}/labels
func (c *Client) AssociateSoundLabel(ctx context.Context, assetID, name string) (*LabelAssociation, error) {
	out, err := fetchJSON[LabelAssociation](c, ctx, rateWrites, "POST /sound-assets/{asset_id}/labels", concurrencyNone, requestSpec{
		method: http.MethodPost,
		path:   "sound-assets/" + assetID + "/labels",
		body:   associateLabelBody{Name: name},
	})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// RemoveSoundLabel removes a label from a sound asset.
//
// DELETE /sound-assets/{asset_id}/labels?name={name}
func (c *Client) RemoveSoundLabel(ctx context.Context, assetID, name string) (*LabelAssociation, error) {
	out, err := fetchJSON[LabelAssociation](c, ctx, rateWrites, "DELETE /sound-assets/{asset_id}/labels", concurrencyNone, requestSpec{
		method: http.MethodDelete,
		path:   "sound-assets/" + assetID + "/labels",
		query:  url.Values{"name": {name}},
	})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// DismissSoundLabelSuggestion dismisses a suggested label for a sound asset.
//
// DELETE /sound-assets/{asset_id}/label-suggestions?name={name}
func (c *Client) DismissSoundLabelSuggestion(ctx context.Context, assetID, name string) (*OkResponse, error) {
	return fetchAck(c, ctx, rateWrites, "DELETE /sound-assets/{asset_id}/label-suggestions", concurrencyNone, requestSpec{
		method: http.MethodDelete,
		path:   "sound-assets/" + assetID + "/label-suggestions",
		query:  url.Values{"name": {name}},
	})
}

// AssociateSavedAnimationLabel adds a label to a saved animation.
//
// POST /saved-animations/{id}/labels
func (c *Client) AssociateSavedAnimationLabel(ctx context.Context, id, name string) (*LabelAssociation, error) {
	out, err := fetchJSON[LabelAssociation](c, ctx, rateWrites, "POST /saved-animations/{id}/labels", concurrencyNone, requestSpec{
		method: http.MethodPost,
		path:   "saved-animations/" + id + "/labels",
		body:   associateLabelBody{Name: name},
	})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// RemoveSavedAnimationLabel removes a label from a saved animation.
//
// DELETE /saved-animations/{id}/labels?name={name}
func (c *Client) RemoveSavedAnimationLabel(ctx context.Context, id, name string) (*LabelAssociation, error) {
	out, err := fetchJSON[LabelAssociation](c, ctx, rateWrites, "DELETE /saved-animations/{id}/labels", concurrencyNone, requestSpec{
		method: http.MethodDelete,
		path:   "saved-animations/" + id + "/labels",
		query:  url.Values{"name": {name}},
	})
	if err != nil {
		return nil, err
	}
	return &out, nil
}
