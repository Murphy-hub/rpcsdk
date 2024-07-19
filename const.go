package rpcsdk

const VERSION = "0.0.1"
const UserAgent = "GO-CLIENT"

var AllowMethods = []string{"get", "delete", "head", "options", "patch", "post", "put"}

const ContentTypeForm = "application/x-www-form-urlencoded"
const ContentTypeJson = "application/json"
const FromAppidKey = "from_appid"
const AccountIdKey = "account_id"
const RequestTraceKey = "X-Request-Id"
const YxtTraceKey = "traceId"

const (
	SelfAppidKey = "x-appid" // X-Appid HTTP_X_APPID
)
