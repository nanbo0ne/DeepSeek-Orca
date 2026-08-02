package main

import (
	"testing"

	"deepseek-orca/internal/event"
)

type mapBackedSink struct{ values map[string]string }

func (mapBackedSink) Emit(event.Event) {}

func TestConversationBrokerRegistersCompleteToolSet(t *testing.T) {
	broker := NewConversationBroker(NewApp())
	tools := broker.Tools("automation-tab", "automation-topic")
	want := []string{"conversation_list", "conversation_read", "conversation_dispatch", "conversation_wait", "conversation_status", "conversation_cancel", "conversation_create"}
	if len(tools) != len(want) {
		t.Fatalf("tools = %d, want %d", len(tools), len(want))
	}
	for i, name := range want {
		if tools[i].Name() != name {
			t.Fatalf("tool %d = %q, want %q", i, tools[i].Name(), name)
		}
	}
}

func TestConversationBrokerTaskOwnershipAndCancel(t *testing.T) {
	broker := NewConversationBroker(NewApp())
	task := &DispatchTask{ID: "task-1", Status: "running", sourceTabID: "source-a", done: make(chan struct{})}
	broker.tasks[task.ID] = task
	if _, err := broker.Status("source-b", task.ID); err == nil {
		t.Fatal("another automation conversation must not inspect the task")
	}
	got, err := broker.Cancel("source-a", task.ID)
	if err != nil || got.Status != "cancelled" {
		t.Fatalf("cancel = %+v, err=%v", got, err)
	}
	select {
	case <-task.done:
	default:
		t.Fatal("cancel must close the task completion channel")
	}
}

func TestConversationBrokerCancelActiveOnlyStopsOwnedTasks(t *testing.T) {
	broker := NewConversationBroker(NewApp())
	owned := &DispatchTask{ID: "owned", Status: "queued", sourceTabID: "source-a", done: make(chan struct{})}
	other := &DispatchTask{ID: "other", Status: "running", sourceTabID: "source-b", done: make(chan struct{})}
	broker.tasks[owned.ID] = owned
	broker.tasks[other.ID] = other
	if !broker.CancelActive("source-a") || owned.Status != "cancelled" {
		t.Fatal("owned active task was not cancelled")
	}
	if other.Status != "running" {
		t.Fatal("another automation conversation's task was cancelled")
	}
}

func TestConversationBrokerSinkUnregisterUsesToken(t *testing.T) {
	broker := NewConversationBroker(NewApp())
	unregisterOld := broker.RegisterSourceSink("source", mapBackedSink{values: map[string]string{"old": "1"}})
	unregisterNew := broker.RegisterSourceSink("source", mapBackedSink{values: map[string]string{"new": "1"}})
	unregisterOld()
	if broker.sourceSinks["source"].sink == nil {
		t.Fatal("stale unregister removed the newer sink")
	}
	unregisterNew()
	if _, ok := broker.sourceSinks["source"]; ok {
		t.Fatal("current unregister did not remove sink")
	}
}
