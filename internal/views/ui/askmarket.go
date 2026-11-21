package ui

import (
	"context"
	"fmt"
	"slices"
)

// The sentence the system prompt gains when the market price tool is left out
// of a route that may not receive public data (D-226), so that a question for a
// price gets the true reason and no price from memory. It is public text (T2).
// An install with no market-data connector, a demo included, adds nothing: the
// tool is not there, and the prompt's rule that every figure comes from a tool
// already keeps the model from quoting a price it was not given.
const (
	// noMarketRouteNotice is for a route none of whose models the owner cleared
	// for public data, T2 (D-206): the tool is left out, and the box answers
	// that prices need a model cleared for market data.
	noMarketRouteNotice = "Market prices are not available to you on this route: prices need a model cleared for market data, and the owner has " +
		"not cleared this one. If you are asked for the price of an instrument, say exactly that prices need a model cleared for market data " +
		"and that the owner can allow it in Settings, and never quote a price from memory."
)

// marketTools returns the names of the tools of tools that fetch market data
// through the owner's connector (answer.Tool.Market).
func marketTools(tools []askTool) []string {
	var names []string
	for i := range tools {
		if tools[i].Market {
			names = append(names, tools[i].Name)
		}
	}
	return names
}

// offer returns the tools the model of a question is told, and the sentence
// the system prompt gains when one is left out (D-226). A market tool is
// offered only when the install has a market-data connector it can ask, which
// the read API's settings say (a demo has none, whatever is installed), and
// the router allows a model of the route to receive its data together with the
// question's: the question is T0 and the tool's card T2 (D-206), so the route
// needs a model allowed for both. A tool left out is not a refusal the model
// can meet: it is not in the list and is no tool of the conversation.
// offering is what the model of a question is told of the tools: the tools it
// may call and a sentence for the system prompt, when there is one.
type offering struct {
	tools  []askTool
	notice string
}

func (a *modelAnswerer) offer(ctx context.Context, sess session) (offering, error) {
	market := marketTools(a.tools)
	if len(market) == 0 {
		return offering{tools: a.tools}, nil
	}
	without := func() []askTool {
		return slices.DeleteFunc(slices.Clone(a.tools), func(t askTool) bool { return t.Market })
	}
	if !sess.Settings.GetMarketQuotes() {
		return offering{tools: without()}, nil
	}
	ok, err := a.model.permits(ctx, sess, []string{tierT0, tierT2})
	if err != nil {
		return offering{}, fmt.Errorf("asking whether the route may receive market data: %w", err)
	}
	if !ok {
		return offering{tools: without(), notice: noMarketRouteNotice}, nil
	}
	return offering{tools: a.tools}, nil
}

// toolOffered returns the tool called name among the tools the question's
// model was told.
func (c *askConversation) toolOffered(name string) (askTool, bool) {
	i := slices.IndexFunc(c.tools, func(t askTool) bool { return t.Name == name })
	if i < 0 {
		return askTool{}, false
	}
	return c.tools[i], true
}

// withNotice returns the system prompt with notice after it, when there is one.
func withNotice(prompt, notice string) string {
	if notice == "" {
		return prompt
	}
	return prompt + " " + notice
}
