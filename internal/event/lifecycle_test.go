package event

import "testing"

func TestLifecycleAnnotatesCommittedAnswerAndTurnOutcome(t *testing.T) {
	var got []Event
	sink := Lifecycle(FuncSink(func(e Event) { got = append(got, e) }))
	sink.Emit(Event{Kind: TurnStarted})
	sink.Emit(Event{Kind: Text, Text: "working"})
	sink.Emit(Event{Kind: Message, Text: "working"})
	sink.Emit(Event{Kind: AnswerCommitted, Text: "working"})
	sink.Emit(Event{Kind: TurnDone, Outcome: TurnOutcomeSuccess})

	var started, message, committed, done Event
	for _, e := range got {
		switch e.Kind {
		case TurnStarted:
			started = e
		case Message:
			message = e
		case AnswerCommitted:
			committed = e
		case TurnDone:
			done = e
		}
	}
	if started.TurnID == "" || message.TurnID != started.TurnID || done.TurnID != started.TurnID {
		t.Fatalf("turn ids = start %q message %q done %q", started.TurnID, message.TurnID, done.TurnID)
	}
	if message.MessageID == "" || committed.FinalMessageID != message.MessageID || done.FinalMessageID != message.MessageID {
		t.Fatalf("message ids = message %q committed %q done %q", message.MessageID, committed.FinalMessageID, done.FinalMessageID)
	}
	if message.ItemID == "" || committed.FinalItemID != message.ItemID || done.FinalItemID != message.ItemID {
		t.Fatalf("item ids = message %q committed %q done %q", message.ItemID, committed.FinalItemID, done.FinalItemID)
	}
	if done.Outcome != TurnOutcomeSuccess {
		t.Fatalf("outcome = %q, want success", done.Outcome)
	}
	var deltas int
	for _, e := range got {
		if e.Kind == ItemDelta && e.TurnID == started.TurnID && e.ItemID == message.ItemID {
			deltas++
		}
	}
	if deltas == 0 {
		t.Fatal("streaming message did not emit a canonical item_delta")
	}
}

func TestLifecycleKeepsIntermediateAndFinalMessagesDistinct(t *testing.T) {
	var messages []Event
	sink := Lifecycle(FuncSink(func(e Event) {
		if e.Kind == Message || e.Kind == AnswerCommitted {
			messages = append(messages, e)
		}
	}))
	sink.Emit(Event{Kind: TurnStarted})
	sink.Emit(Event{Kind: Text, Text: "checking"})
	sink.Emit(Event{Kind: Message, Text: "checking"})
	sink.Emit(Event{Kind: ToolDispatch, Tool: Tool{ID: "tool-1", Name: "read_file"}})
	sink.Emit(Event{Kind: ToolResult, Tool: Tool{ID: "tool-1", Name: "read_file"}})
	sink.Emit(Event{Kind: Text, Text: "done"})
	sink.Emit(Event{Kind: Message, Text: "done"})
	sink.Emit(Event{Kind: AnswerCommitted, Text: "done"})

	if len(messages) != 3 {
		t.Fatalf("message lifecycle events = %d, want 3", len(messages))
	}
	if messages[0].MessageID == messages[1].MessageID {
		t.Fatal("intermediate and final provider rounds shared a message id")
	}
	if messages[2].FinalMessageID != messages[1].MessageID {
		t.Fatalf("committed %q, want final %q", messages[2].FinalMessageID, messages[1].MessageID)
	}
}

func TestLifecycleDoesNotReportSuccessWithoutCommittedAnswer(t *testing.T) {
	var done Event
	sink := Lifecycle(FuncSink(func(e Event) {
		if e.Kind == TurnDone {
			done = e
		}
	}))
	sink.Emit(Event{Kind: TurnStarted})
	sink.Emit(Event{Kind: Text, Text: "progress only"})
	sink.Emit(Event{Kind: Message, Text: "progress only"})
	sink.Emit(Event{Kind: TurnDone, Outcome: TurnOutcomeSuccess})

	if done.Outcome != TurnOutcomeInterrupted {
		t.Fatalf("outcome = %q, want interrupted", done.Outcome)
	}
	if done.FinalItemID != "" || done.FinalMessageID != "" {
		t.Fatalf("uncommitted turn exposed final ids: %+v", done)
	}
}
