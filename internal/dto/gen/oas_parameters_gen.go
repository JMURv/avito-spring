// Code generated by ogen, DO NOT EDIT.

package dto

import (
	"net/http"
	"net/url"
	"time"

	"github.com/go-faster/errors"
	"github.com/google/uuid"

	"github.com/ogen-go/ogen/conv"
	"github.com/ogen-go/ogen/middleware"
	"github.com/ogen-go/ogen/ogenerrors"
	"github.com/ogen-go/ogen/uri"
	"github.com/ogen-go/ogen/validate"
)

// PvzGetParams is parameters of GET /pvz operation.
type PvzGetParams struct {
	// Начальная дата диапазона.
	StartDate OptDateTime
	// Конечная дата диапазона.
	EndDate OptDateTime
	// Номер страницы.
	Page OptInt
	// Количество элементов на странице.
	Limit OptInt
}

func unpackPvzGetParams(packed middleware.Parameters) (params PvzGetParams) {
	{
		key := middleware.ParameterKey{
			Name: "startDate",
			In:   "query",
		}
		if v, ok := packed[key]; ok {
			params.StartDate = v.(OptDateTime)
		}
	}
	{
		key := middleware.ParameterKey{
			Name: "endDate",
			In:   "query",
		}
		if v, ok := packed[key]; ok {
			params.EndDate = v.(OptDateTime)
		}
	}
	{
		key := middleware.ParameterKey{
			Name: "page",
			In:   "query",
		}
		if v, ok := packed[key]; ok {
			params.Page = v.(OptInt)
		}
	}
	{
		key := middleware.ParameterKey{
			Name: "limit",
			In:   "query",
		}
		if v, ok := packed[key]; ok {
			params.Limit = v.(OptInt)
		}
	}
	return params
}

func decodePvzGetParams(args [0]string, argsEscaped bool, r *http.Request) (params PvzGetParams, _ error) {
	q := uri.NewQueryDecoder(r.URL.Query())
	// Decode query: startDate.
	if err := func() error {
		cfg := uri.QueryParameterDecodingConfig{
			Name:    "startDate",
			Style:   uri.QueryStyleForm,
			Explode: true,
		}

		if err := q.HasParam(cfg); err == nil {
			if err := q.DecodeParam(cfg, func(d uri.Decoder) error {
				var paramsDotStartDateVal time.Time
				if err := func() error {
					val, err := d.DecodeValue()
					if err != nil {
						return err
					}

					c, err := conv.ToDateTime(val)
					if err != nil {
						return err
					}

					paramsDotStartDateVal = c
					return nil
				}(); err != nil {
					return err
				}
				params.StartDate.SetTo(paramsDotStartDateVal)
				return nil
			}); err != nil {
				return err
			}
		}
		return nil
	}(); err != nil {
		return params, &ogenerrors.DecodeParamError{
			Name: "startDate",
			In:   "query",
			Err:  err,
		}
	}
	// Decode query: endDate.
	if err := func() error {
		cfg := uri.QueryParameterDecodingConfig{
			Name:    "endDate",
			Style:   uri.QueryStyleForm,
			Explode: true,
		}

		if err := q.HasParam(cfg); err == nil {
			if err := q.DecodeParam(cfg, func(d uri.Decoder) error {
				var paramsDotEndDateVal time.Time
				if err := func() error {
					val, err := d.DecodeValue()
					if err != nil {
						return err
					}

					c, err := conv.ToDateTime(val)
					if err != nil {
						return err
					}

					paramsDotEndDateVal = c
					return nil
				}(); err != nil {
					return err
				}
				params.EndDate.SetTo(paramsDotEndDateVal)
				return nil
			}); err != nil {
				return err
			}
		}
		return nil
	}(); err != nil {
		return params, &ogenerrors.DecodeParamError{
			Name: "endDate",
			In:   "query",
			Err:  err,
		}
	}
	// Set default value for query: page.
	{
		val := int(1)
		params.Page.SetTo(val)
	}
	// Decode query: page.
	if err := func() error {
		cfg := uri.QueryParameterDecodingConfig{
			Name:    "page",
			Style:   uri.QueryStyleForm,
			Explode: true,
		}

		if err := q.HasParam(cfg); err == nil {
			if err := q.DecodeParam(cfg, func(d uri.Decoder) error {
				var paramsDotPageVal int
				if err := func() error {
					val, err := d.DecodeValue()
					if err != nil {
						return err
					}

					c, err := conv.ToInt(val)
					if err != nil {
						return err
					}

					paramsDotPageVal = c
					return nil
				}(); err != nil {
					return err
				}
				params.Page.SetTo(paramsDotPageVal)
				return nil
			}); err != nil {
				return err
			}
			if err := func() error {
				if value, ok := params.Page.Get(); ok {
					if err := func() error {
						if err := (validate.Int{
							MinSet:        true,
							Min:           1,
							MaxSet:        false,
							Max:           0,
							MinExclusive:  false,
							MaxExclusive:  false,
							MultipleOfSet: false,
							MultipleOf:    0,
						}).Validate(int64(value)); err != nil {
							return errors.Wrap(err, "int")
						}
						return nil
					}(); err != nil {
						return err
					}
				}
				return nil
			}(); err != nil {
				return err
			}
		}
		return nil
	}(); err != nil {
		return params, &ogenerrors.DecodeParamError{
			Name: "page",
			In:   "query",
			Err:  err,
		}
	}
	// Set default value for query: limit.
	{
		val := int(10)
		params.Limit.SetTo(val)
	}
	// Decode query: limit.
	if err := func() error {
		cfg := uri.QueryParameterDecodingConfig{
			Name:    "limit",
			Style:   uri.QueryStyleForm,
			Explode: true,
		}

		if err := q.HasParam(cfg); err == nil {
			if err := q.DecodeParam(cfg, func(d uri.Decoder) error {
				var paramsDotLimitVal int
				if err := func() error {
					val, err := d.DecodeValue()
					if err != nil {
						return err
					}

					c, err := conv.ToInt(val)
					if err != nil {
						return err
					}

					paramsDotLimitVal = c
					return nil
				}(); err != nil {
					return err
				}
				params.Limit.SetTo(paramsDotLimitVal)
				return nil
			}); err != nil {
				return err
			}
			if err := func() error {
				if value, ok := params.Limit.Get(); ok {
					if err := func() error {
						if err := (validate.Int{
							MinSet:        true,
							Min:           1,
							MaxSet:        true,
							Max:           30,
							MinExclusive:  false,
							MaxExclusive:  false,
							MultipleOfSet: false,
							MultipleOf:    0,
						}).Validate(int64(value)); err != nil {
							return errors.Wrap(err, "int")
						}
						return nil
					}(); err != nil {
						return err
					}
				}
				return nil
			}(); err != nil {
				return err
			}
		}
		return nil
	}(); err != nil {
		return params, &ogenerrors.DecodeParamError{
			Name: "limit",
			In:   "query",
			Err:  err,
		}
	}
	return params, nil
}

// PvzPvzIdCloseLastReceptionPostParams is parameters of POST /pvz/{pvzId}/close_last_reception operation.
type PvzPvzIdCloseLastReceptionPostParams struct {
	PvzId uuid.UUID
}

func unpackPvzPvzIdCloseLastReceptionPostParams(packed middleware.Parameters) (params PvzPvzIdCloseLastReceptionPostParams) {
	{
		key := middleware.ParameterKey{
			Name: "pvzId",
			In:   "path",
		}
		params.PvzId = packed[key].(uuid.UUID)
	}
	return params
}

func decodePvzPvzIdCloseLastReceptionPostParams(args [1]string, argsEscaped bool, r *http.Request) (params PvzPvzIdCloseLastReceptionPostParams, _ error) {
	// Decode path: pvzId.
	if err := func() error {
		param := args[0]
		if argsEscaped {
			unescaped, err := url.PathUnescape(args[0])
			if err != nil {
				return errors.Wrap(err, "unescape path")
			}
			param = unescaped
		}
		if len(param) > 0 {
			d := uri.NewPathDecoder(uri.PathDecoderConfig{
				Param:   "pvzId",
				Value:   param,
				Style:   uri.PathStyleSimple,
				Explode: false,
			})

			if err := func() error {
				val, err := d.DecodeValue()
				if err != nil {
					return err
				}

				c, err := conv.ToUUID(val)
				if err != nil {
					return err
				}

				params.PvzId = c
				return nil
			}(); err != nil {
				return err
			}
		} else {
			return validate.ErrFieldRequired
		}
		return nil
	}(); err != nil {
		return params, &ogenerrors.DecodeParamError{
			Name: "pvzId",
			In:   "path",
			Err:  err,
		}
	}
	return params, nil
}

// PvzPvzIdDeleteLastProductPostParams is parameters of POST /pvz/{pvzId}/delete_last_product operation.
type PvzPvzIdDeleteLastProductPostParams struct {
	PvzId uuid.UUID
}

func unpackPvzPvzIdDeleteLastProductPostParams(packed middleware.Parameters) (params PvzPvzIdDeleteLastProductPostParams) {
	{
		key := middleware.ParameterKey{
			Name: "pvzId",
			In:   "path",
		}
		params.PvzId = packed[key].(uuid.UUID)
	}
	return params
}

func decodePvzPvzIdDeleteLastProductPostParams(args [1]string, argsEscaped bool, r *http.Request) (params PvzPvzIdDeleteLastProductPostParams, _ error) {
	// Decode path: pvzId.
	if err := func() error {
		param := args[0]
		if argsEscaped {
			unescaped, err := url.PathUnescape(args[0])
			if err != nil {
				return errors.Wrap(err, "unescape path")
			}
			param = unescaped
		}
		if len(param) > 0 {
			d := uri.NewPathDecoder(uri.PathDecoderConfig{
				Param:   "pvzId",
				Value:   param,
				Style:   uri.PathStyleSimple,
				Explode: false,
			})

			if err := func() error {
				val, err := d.DecodeValue()
				if err != nil {
					return err
				}

				c, err := conv.ToUUID(val)
				if err != nil {
					return err
				}

				params.PvzId = c
				return nil
			}(); err != nil {
				return err
			}
		} else {
			return validate.ErrFieldRequired
		}
		return nil
	}(); err != nil {
		return params, &ogenerrors.DecodeParamError{
			Name: "pvzId",
			In:   "path",
			Err:  err,
		}
	}
	return params, nil
}
