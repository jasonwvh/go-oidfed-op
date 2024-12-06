package fedentities

import (
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/zachmann/go-oidfed/pkg"
	"github.com/zachmann/go-oidfed/pkg/apimodel"
	"github.com/zachmann/go-oidfed/pkg/constants"
	"github.com/zachmann/go-oidfed/pkg/unixtime"
)

// AddResolveEndpoint adds a resolve endpoint
func (fed *FedEntity) AddResolveEndpoint(endpoint EndpointConf) {
	fed.Metadata.FederationEntity.FederationResolveEndpoint = endpoint.ValidateURL(fed.FederationEntity.EntityID)
	if endpoint.Path == "" {
		return
	}
	fed.server.Get(
		endpoint.Path, func(ctx *fiber.Ctx) error {
			var req apimodel.ResolveRequest
			if err := ctx.QueryParser(&req); err != nil {
				ctx.Status(fiber.StatusBadRequest)
				return ctx.JSON(pkg.ErrorInvalidRequest("could not parse request parameters: " + err.Error()))
			}
			if len(req.Anchor) == 0 {
				ctx.Status(fiber.StatusBadRequest)
				return ctx.JSON(pkg.ErrorInvalidRequest("required parameter 'anchor' not given"))
			}
			if req.Subject == "" {
				ctx.Status(fiber.StatusBadRequest)
				return ctx.JSON(pkg.ErrorInvalidRequest("required parameter 'sub' not given"))
			}
			resolver := pkg.TrustResolver{
				TrustAnchors:   pkg.NewTrustAnchorsFromEntityIDs(req.Anchor...),
				StartingEntity: req.Subject,
				Types:          req.EntityTypes,
			}
			chains := resolver.ResolveToValidChains()
			if len(chains) == 0 {
				ctx.Status(fiber.StatusNotFound)
				return ctx.JSON(pkg.ErrorInvalidRequest("no valid trust path between sub and anchor found"))
			}
			selectedChain := chains.Filter(pkg.TrustChainsFilterMinPathLength)[0]
			metadata, _ := selectedChain.Metadata()
			// err cannot be != nil, since ResolveToValidChains only gives chains with valid metadata
			leaf := selectedChain[0]
			ta := selectedChain[len(selectedChain)-1]
			res := pkg.ResolveResponse{
				Issuer:     fed.FederationEntity.EntityID,
				Subject:    req.Subject,
				IssuedAt:   unixtime.Unixtime{Time: time.Now()},
				ExpiresAt:  selectedChain.ExpiresAt(),
				ResolveResponsePayload: pkg.ResolveResponsePayload{
				Metadata:   metadata,
				TrustMarks: leaf.TrustMarks.VerifiedFederation(&ta.EntityStatementPayload),
				TrustChain: selectedChain.Messages(),
				},
			}
			jwt, err := fed.GeneralJWTSigner.ResolveResponseSigner().JWT(res)
			if err != nil {
				ctx.Status(fiber.StatusInternalServerError)
				return ctx.JSON(pkg.ErrorServerError(err.Error()))
			}
			ctx.Set(fiber.HeaderContentType, constants.ContentTypeResolveResponse)
			return ctx.Send(jwt)
		},
	)
}
