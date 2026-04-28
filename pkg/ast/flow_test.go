package ast

import "testing"

func TestValidateFlows_MissingID(t *testing.T) {
    plan := &TrackingPlan{
        Events: map[string]Event{"Login": {Category: "INTERACTION", EntityType: "User"}},
        Flows: []Flow{{
            ID: "",
            Steps: []FlowStep{{State: "LoginPage", Event: "Login", Triggers: []string{"DIRECT_LOAD"}}},
        }},
    }
    if err := plan.ValidateFlows(); err == nil {
        t.Errorf("expected error for missing flow ID, got nil")
    }
}

func TestValidateFlows_UnknownEvent(t *testing.T) {
    plan := &TrackingPlan{
        Events: map[string]Event{"Login": {Category: "INTERACTION", EntityType: "User"}},
        Flows: []Flow{{
            ID: "flow1",
            Steps: []FlowStep{{State: "SignupPage", Event: "Signup", Triggers: []string{"DIRECT_LOAD"}}},
        }},
    }
    if err := plan.ValidateFlows(); err == nil {
        t.Errorf("expected error for unknown event in flow step, got nil")
    }
}

func TestValidateFlows_UnknownTrigger(t *testing.T) {
    plan := &TrackingPlan{
        Events: map[string]Event{"Login": {Category: "INTERACTION", EntityType: "User"}},
        Flows: []Flow{{
            ID: "flow1",
            Steps: []FlowStep{{State: "LoginPage", Event: "Login", Triggers: []string{"MAGIC_BUTTON"}}},
        }},
    }
    if err := plan.ValidateFlows(); err == nil {
        t.Errorf("expected error for unknown trigger, got nil")
    }
}

func TestValidateFlows_OK(t *testing.T) {
    plan := &TrackingPlan{
        Events: map[string]Event{"Login": {Category: "INTERACTION", EntityType: "User"}},
        Flows: []Flow{{
            ID: "flow1",
            Steps: []FlowStep{{State: "LoginPage", Event: "Login", Triggers: []string{"DIRECT_LOAD", "UI_BACK"}}},
        }},
    }
    if err := plan.ValidateFlows(); err != nil {
        t.Fatalf("expected no error, got %v", err)
    }
}
