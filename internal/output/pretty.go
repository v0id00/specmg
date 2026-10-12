package output

import "github.com/v0id00/specmg/internal/spec"

// statusIcons are badges for human (pretty) output.
var statusIcons = map[string]string{
	spec.StatusDraft:       "🟡",
	spec.StatusInReview:    "🔵",
	spec.StatusApproved:    "🟢",
	spec.StatusImplemented: "🟣",
	spec.StatusVerified:    "✅",
	spec.StatusSuperseded:  "⚫",
	spec.StatusRejected:    "🔴",
}

// StatusBadge returns "icon status" for human output; the plain status when
// unknown.
func StatusBadge(status string) string {
	if icon, ok := statusIcons[status]; ok {
		return icon + " " + status
	}
	return status
}

// Item icons for requirement/criterion/question markers.
const (
	IconReq   = "🔹" // requirement
	IconACYes = "☑" // checked acceptance criterion
	IconACNo  = "☐" // unchecked acceptance criterion
	IconQ     = "❓" // open question
	IconBlock = "❗" // blocking question
	IconDone  = "✅" // resolved question / passing check
)

// CriterionIcon returns the checkbox icon for a criterion.
func CriterionIcon(done bool) string {
	if done {
		return IconACYes
	}
	return IconACNo
}

// QuestionIcon returns the marker for a question state.
func QuestionIcon(blocking, resolved bool) string {
	switch {
	case resolved:
		return IconDone
	case blocking:
		return IconBlock
	default:
		return IconQ
	}
}
