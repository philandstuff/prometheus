package promlint

import (
	"fmt"
	"os"
	"strings"

	"github.com/prometheus/prometheus/promql"
)

// types of instant vectors:

//   counter
//   gauge

// extensive values (ie values that do change when sample rate changes)

//   rate
//

// types enforced by `promtool check metrics`:
// counters: suffix _total required
// histograms: suffix _bucket exclusivex
// histograms and summaries: suffix _count and _sum exclusive

// types of some functions
// rate: Range Counter -> 
// range selector: s -> Range s

// functions that take specific inputs
// counter range vector: rate, irate, increase, resets
// gauge range vector: delta, deriv, holt_winters, idelta, predict_linear

func isCounterFunc(node promql.Node) bool {
	switch n := node.(type) {
	case *promql.Call:
		switch n.Func.Name {
		case "increase","irate","rate","resets":
			return true
		}
	}
	return false
}

func isCounterMetricName(name string) bool{
	return strings.HasSuffix(name, "_total")
}

func isCounter(node promql.Node) bool {
	var name string
	switch n := node.(type) {
	case *promql.VectorSelector:
		name = n.Name
	case *promql.MatrixSelector:
		name = n.Name
	default:
		return false
	}
	return isCounterMetricName(name)
}

func CheckNode(node promql.Node, path []promql.Node) error {
	switch n := node.(type) {
	case *promql.Call:
		if isCounterFunc(n) {
			arg := n.Args[0]
			if !isCounter(arg) {
				fmt.Fprintln(os.Stderr, "error: function", n.Func.Name, "must be called on a counter selector in", node)
			}
		}
	case *promql.MatrixSelector:
		if isCounterMetricName(n.Name) {
			if len(path) == 0 {
				return nil
			}
			parent := path[len(path)-2]
			if !isCounterFunc(parent) {
				fmt.Fprintln(os.Stderr, "error: counter selector", n, "used in expression without first passing through rate(), irate(), increase() or resets()")
			}
		}
	case *promql.VectorSelector:
		if isCounterMetricName(n.Name) {
			if len(path) == 0 {
				return nil
			}
			parent := path[len(path)-2]
			if !isCounterFunc(parent) {
				fmt.Fprintln(os.Stderr, "error: counter selector", n, "used in expression without first passing through rate(), irate(), increase() or resets()")
			}
		}
	}
	return nil
}

func CheckExpr(expr promql.Expr) error {
	promql.Inspect(expr,CheckNode)
	return nil
}
