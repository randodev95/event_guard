package generator

import (
	"bytes"
	"github.com/randodev95/event_guard/pkg/ast"
	"html/template"
)

const htmlTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>EventGuard: Tracking Plan</title>
    <style>
        :root {
            --primary: #6366f1;
            --bg: #0f172a;
            --card: #1e293b;
            --text: #f8fafc;
            --muted: #94a3b8;
            --accent: #22d3ee;
        }
        body {
            font-family: 'Inter', system-ui, sans-serif;
            background: var(--bg);
            color: var(--text);
            line-height: 1.6;
            margin: 0;
            padding: 2rem;
        }
        .container {
            max-width: 1000px;
            margin: 0 auto;
        }
        header {
            margin-bottom: 3rem;
            border-bottom: 1px solid var(--card);
            padding-bottom: 1rem;
        }
        h1 {
            font-size: 2.5rem;
            background: linear-gradient(to right, var(--primary), var(--accent));
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
            margin: 0;
        }
        .search-container {
            margin-bottom: 2rem;
        }
        input {
            width: 100%;
            padding: 1rem;
            background: var(--card);
            border: 1px solid var(--muted);
            border-radius: 8px;
            color: var(--text);
            font-size: 1rem;
        }
        .event-card {
            background: var(--card);
            border-radius: 12px;
            padding: 1.5rem;
            margin-bottom: 1.5rem;
            transition: transform 0.2s;
        }
        .event-card:hover {
            transform: translateY(-4px);
            box-shadow: 0 10px 20px rgba(0,0,0,0.3);
        }
        .event-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 1rem;
        }
        .event-name {
            font-size: 1.25rem;
            font-weight: 600;
            color: var(--accent);
        }
        .property-table {
            width: 100%;
            border-collapse: collapse;
        }
        .property-table th {
            text-align: left;
            color: var(--muted);
            font-size: 0.875rem;
            text-transform: uppercase;
            padding: 0.5rem;
            border-bottom: 1px solid var(--muted);
        }
        .property-table td {
            padding: 0.75rem 0.5rem;
            border-bottom: 1px solid rgba(255,255,255,0.05);
        }
        .badge {
            padding: 0.25rem 0.5rem;
            border-radius: 4px;
            font-size: 0.75rem;
            font-weight: 600;
        }
        .badge-required { background: #ef4444; color: white; }
        .badge-optional { background: var(--muted); color: var(--bg); }
    </style>
</head>
<body>
    <div class="container">
        <header>
            <h1>EventGuard Live Docs</h1>
            <p style="color: var(--muted)">Deterministic Tracking Plan Documentation</p>
        </header>

        <div class="search-container">
            <input type="text" id="search" placeholder="Search events or properties..." onkeyup="filter()">
        </div>

        <div id="flows">
            {{range $flowName, $flow := .Flows}}
            <div class="event-card" style="border-left: 4px solid var(--accent)">
                <div class="event-header">
                    <span class="event-name">Flow: {{$flow.Namespace}}</span>
                </div>
                <div style="padding: 1rem 0">
                    {{range $nodeName, $node := $flow.Nodes}}
                    <div style="display: flex; align-items: flex-start; margin-bottom: 1rem">
                        <div style="background: var(--primary); width: 24px; height: 24px; border-radius: 50%; display: flex; align-items: center; justify-content: center; font-size: 0.75rem; margin-right: 1rem; flex-shrink: 0">
                            {{$node.Type}}
                        </div>
                        <div>
                            <div style="font-weight: 600">{{$nodeName}}</div>
                            <div style="font-size: 0.875rem; color: var(--muted)">
                                {{if $node.Event}}Trigger: {{$node.Event}}{{end}}
                                {{if $node.ListenFor}}Listen: {{$node.ListenFor}}{{end}}
                            </div>
                        </div>
                    </div>
                    {{end}}
                </div>
            </div>
            {{end}}
        </div>

        <div id="events">
            {{range .ResolvedEvents}}
            <div class="event-card">
                <div class="event-header">
                    <span class="event-name">{{.Name}}</span>
                </div>
                <div style="color: var(--muted); margin-bottom: 1rem">{{.Description}}</div>
                <table class="property-table">
                    <thead>
                        <tr>
                            <th>Property</th>
                            <th>Type</th>
                            <th>Requirement</th>
                        </tr>
                    </thead>
                    <tbody>
                        {{range $propName, $prop := .Properties}}
                        <tr>
                            <td><code>{{$propName}}</code></td>
                            <td><span style="color: var(--primary)">{{$prop.Type}}</span></td>
                            <td>
                                {{if $prop.Required}}
                                <span class="badge badge-required">REQUIRED</span>
                                {{else}}
                                <span class="badge badge-optional">OPTIONAL</span>
                                {{end}}
                            </td>
                        </tr>
                        {{end}}
                    </tbody>
                </table>
            </div>
            {{end}}
        </div>
    </div>
</body>
</html>
`

// HTMLData holds the template data for the HTML documentation generator.
type HTMLData struct {
	Flows          map[string]ast.FlowPlan
	ResolvedEvents []ResolvedEvent
}

// GenerateHTML renders the tracking plan into a premium, interactive HTML documentation page.
func GenerateHTML(plan *ast.TrackingPlan) (string, error) {
	resolved, err := getResolvedEvents(plan)
	if err != nil {
		return "", err
	}

	data := HTMLData{
		Flows:          plan.Flows,
		ResolvedEvents: resolved,
	}

	tmpl, err := template.New("docs").Parse(htmlTemplate)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}
