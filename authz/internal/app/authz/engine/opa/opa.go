package opa

// nolint:lll
//go:generate go-bindata -pkg $GOPACKAGE -o policy.bindata.go -ignore .*_test.rego -ignore Makefile -ignore README\.md policy/...

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-kit/kit/log"
	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/rego"
	"github.com/open-policy-agent/opa/storage"
	"github.com/open-policy-agent/opa/storage/inmem"
	"github.com/pkg/errors"
)

// State wraps the state of OPA we need to track
type State struct {
	log         log.Logger
	store       storage.Store
	queries     map[string]ast.Body
	compiler    *ast.Compiler
	modules     map[string]*ast.Module
	partialAuth rego.PartialResult
}

// this needs to match the hardcoded OPA policy document we've put in place
const (
	authzQuery = "data.rbac.authz.allow"
)

// OptFunc is the type of functional options to be passed to New()
type OptFunc func(*State)

// New initializes a fresh OPA state, using the default, hardcoded OPA policy
// from policy/authz*.rego unless overridden via an opa.OptFunc.
func New(ctx context.Context, l log.Logger, opts ...OptFunc) (*State, error) {
	authzQueryParsed, err := ast.ParseBody(authzQuery)
	if err != nil {
		return nil, errors.Wrapf(err, "parse query %q", authzQuery)
	}

	s := State{
		log:   l,
		store: inmem.New(),
		queries: map[string]ast.Body{
			authzQuery: authzQueryParsed,
		},
	}
	for _, opt := range opts {
		opt(&s)
	}

	if err := s.initModules(); err != nil {
		return nil, errors.Wrap(err, "init OPA modules")
	}
	return &s, nil
}

// initModules parses the rego files that have been compiled-in and stores the
// result, to be used in initPartialResult.
func (s *State) initModules() error {
	if len(s.modules) == 0 {
		mods := map[string]*ast.Module{}
		for _, name := range AssetNames() {
			if !strings.HasSuffix(name, ".rego") {
				continue // skip this, whatever has been compiled-in here
			}
			parsed, err := ast.ParseModule(name, string(MustAsset(name)))
			if err != nil {
				return errors.Wrapf(err, "parse policy file %q", name)
			}
			mods[name] = parsed
		}
		s.modules = mods
	}

	// this compiler is for the ad-hoc queries (those *not* having partial results prepped)
	compiler, err := s.newCompiler()
	if err != nil {
		return errors.Wrap(err, "init compiler")
	}
	s.compiler = compiler
	return nil
}

func (s *State) newCompiler() (*ast.Compiler, error) {
	compiler := ast.NewCompiler()
	compiler.Compile(s.modules)
	if compiler.Failed() {
		return nil, errors.Wrap(compiler.Errors, "compile modules")
	}

	return compiler, nil
}

// SetUserRolesAndPermissions replaces OPA's data with a new set of user role, and resets the
// partial evaluation cache
func (s *State) SetUserRolesAndPermissions(ctx context.Context, userRoles map[string]interface{}) error {
	s.store = inmem.NewFromObject(userRoles)
	return s.initPartialResult(ctx)
}

// initPartialResult allows caching things that don't change among multiple
// query evaluations. We don't bother for the pairs query, but for
// IsAuthorized(), we want to do as little work per call as possible.
func (s *State) initPartialResult(ctx context.Context) error {
	// Reset compiler to avoid state issues
	compiler, err := s.newCompiler()
	if err != nil {
		return err
	}

	r := rego.New(
		rego.ParsedQuery(s.queries[authzQuery]),
		rego.Compiler(compiler),
		rego.Store(s.store),
	)
	pr, err := r.PartialResult(ctx)
	if err != nil {
		return errors.Wrap(err, "partial eval")
	}
	s.partialAuth = pr
	return nil
}

// IsAuthorized evaluates whether a given [domainId, subject, resource, action] tuple
// is authorized given the service's state
func (s *State) IsAuthorized(
	ctx context.Context,
	user string,
	path string,
	method string) (bool, error) {

	input := ast.NewObject(
		[2]*ast.Term{ast.NewTerm(ast.String("user")), ast.NewTerm(ast.String(user))},
		[2]*ast.Term{ast.NewTerm(ast.String("path")), ast.NewTerm(ast.String(path))},
		[2]*ast.Term{ast.NewTerm(ast.String("method")), ast.NewTerm(ast.String(method))},
	)
	resultSet, err := s.partialAuth.Rego(rego.ParsedInput(input)).Eval(ctx)
	if err != nil {
		return false, &ErrEvaluation{e: err}
	}

	switch len(resultSet) {
	case 0:
		return false, nil
	case 1:
		exps := resultSet[0].Expressions
		if len(exps) != 1 {
			return false, &ErrUnexpectedResultExpression{exps: exps}

		}
		return exps[0].Value == true, nil
	default:
		return false, &ErrUnexpectedResultSet{set: resultSet}
	}
}

// ErrUnexpectedResultExpression is returned when one of the result sets
// expressions can't be made sense of
type ErrUnexpectedResultExpression struct {
	exps []*rego.ExpressionValue
}

func (e *ErrUnexpectedResultExpression) Error() string {
	return fmt.Sprintf("unexpected result expressions: %v", e.exps)
}

// ErrUnexpectedResultSet is returned when the result set of an OPA query
// can't be made sense of
type ErrUnexpectedResultSet struct {
	set rego.ResultSet
}

func (e *ErrUnexpectedResultSet) Error() string {
	return fmt.Sprintf("unexpected result set: %v", e.set)
}

// ErrEvaluation is returned when a query evaluation returns an error.
type ErrEvaluation struct {
	e error
}

func (e *ErrEvaluation) Error() string {
	return fmt.Sprintf("error in query evaluation: %s", e.e.Error())
}
