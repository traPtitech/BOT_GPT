package tooling

import (
	"context"
	"testing"

	"github.com/openai/openai-go/v2/packages/param"
)

func TestSpecToResponseTool_MCP(t *testing.T) {
	spec := Spec{
		Kind: ToolKindMCP,
		MCP: &MCPOptions{
			ServerLabel:    "deepwiki",
			ServerURL:      param.NewOpt("https://mcp.deepwiki.com/mcp"),
			ApprovalPolicy: "never",
		},
	}

	tool, err := spec.ToResponseTool()
	if err != nil {
		t.Fatalf("ToResponseTool returned error: %v", err)
	}

	if tool.OfMcp == nil {
		t.Fatalf("expected MCP tool variant, got nil")
	}

	if tool.OfMcp.ServerLabel != spec.MCP.ServerLabel {
		t.Errorf("ServerLabel = %q, want %q", tool.OfMcp.ServerLabel, spec.MCP.ServerLabel)
	}

	if tool.OfMcp.ServerURL != spec.MCP.ServerURL {
		t.Errorf("ServerURL = %q, want %q", tool.OfMcp.ServerURL, spec.MCP.ServerURL)
	}

	approval := tool.OfMcp.RequireApproval.OfMcpToolApprovalSetting
	if !approval.Valid() {
		t.Fatalf("expected approval setting to be set")
	}

	if approval.Value != spec.MCP.ApprovalPolicy {
		t.Errorf("ApprovalPolicy = %q, want %q", approval.Value, spec.MCP.ApprovalPolicy)
	}
}

func TestSpecToResponseTool_Invalid(t *testing.T) {
	spec := Spec{Kind: ToolKindMCP}

	if _, err := spec.ToResponseTool(); err == nil {
		t.Fatalf("expected error when MCP options missing")
	}
}

func TestStaticProviderTools(t *testing.T) {
	specs := []Spec{
		{
			Kind: ToolKindMCP,
			MCP: &MCPOptions{
				ServerLabel:    "deepwiki",
				ServerURL:      param.NewOpt("https://mcp.deepwiki.com/mcp"),
				ApprovalPolicy: "never",
			},
		},
	}

	provider := NewStaticProvider(specs)

	specs[0].MCP.ServerLabel = "modified"

	tools, err := provider.Tools(context.Background())
	if err != nil {
		t.Fatalf("Tools returned error: %v", err)
	}

	if len(tools) != 1 {
		t.Fatalf("expected 1 tool, got %d", len(tools))
	}

	got := tools[0].OfMcp
	if got == nil {
		t.Fatalf("expected MCP tool variant, got nil")
	}

	if got.ServerLabel != "deepwiki" {
		t.Errorf("ServerLabel = %q, want %q", got.ServerLabel, "deepwiki")
	}
}

func TestStaticProviderTools_InvalidSpec(t *testing.T) {
	provider := NewStaticProvider([]Spec{{Kind: ToolKindMCP}})

	if _, err := provider.Tools(context.Background()); err == nil {
		t.Fatalf("expected error for invalid spec")
	}
}

func TestStaticProviderNil(t *testing.T) {
	var provider *StaticProvider

	tools, err := provider.Tools(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if tools != nil {
		t.Fatalf("expected nil tools, got %v", tools)
	}
}

func TestDefaultSpecsProvider(t *testing.T) {
	provider := NewStaticProvider(DefaultSpecs())

	tools, err := provider.Tools(context.Background())
	if err != nil {
		t.Fatalf("Tools returned error: %v", err)
	}

	if len(tools) != 1 {
		t.Fatalf("expected 1 tool, got %d", len(tools))
	}

	if tools[0].OfMcp == nil {
		t.Fatalf("expected MCP tool variant")
	}

	if got := tools[0].OfMcp.ServerLabel; got != "deepwiki" {
		t.Fatalf("ServerLabel = %q, want %q", got, "deepwiki")
	}
}
