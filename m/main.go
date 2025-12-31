package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	"github.com/casbin/casbin/v2/persist/file-adapter"
	casbinUtil "github.com/casbin/casbin/v2/util"
)

// Subscription 数据结构（你可以把这些信息存在 DB）
type Subscription struct {
	UserID    string
	Package   string // "starter","growth","pro","enterprise","trial","backend_*"
	Expiry    time.Time
	IsAdmin   bool
	CreatedBy string // "backend" or "self"
}

// QuotaManager 简单的并发安全配额管理（内存版）
// key 可以是 userID + ":" + resource
type QuotaManager struct {
	mu    sync.Mutex
	store map[string]int64
	// optional: quota limits map[string]int64
	limits map[string]int64
}

func NewQuotaManager() *QuotaManager {
	return &QuotaManager{
		store:  make(map[string]int64),
		limits: make(map[string]int64),
	}
}

// SetLimit 设置配额上限，例如 Screens 剩余次数初值
func (q *QuotaManager) SetLimit(key string, v int64) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.limits[key] = v
	q.store[key] = v
}

// GetRemaining 返回剩余
func (q *QuotaManager) GetRemaining(key string) int64 {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.store[key]
}

// TryConsume 尝试原子消费 count 个配额，成功返回 true
func (q *QuotaManager) TryConsume(key string, count int64) bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	remain := q.store[key]
	if remain >= count {
		q.store[key] = remain - count
		return true
	}
	return false
}

// Restore 恢复（回滚）配额
func (q *QuotaManager) Restore(key string, count int64) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.store[key] += count
	limit := q.limits[key]
	if limit > 0 && q.store[key] > limit {
		q.store[key] = limit
	}
}

// --- 全局演示数据 ---
var (
	enf      *casbin.SyncedEnforcer
	quotaMan *QuotaManager
)

// initEnforcer 从 model.conf 和 policy.csv 初始化
func initEnforcer() error {
	m, err := model.NewModelFromFile("/Users/wpeng/Projects/golang/src/kit/m/model.conf")
	if err != nil {
		return err
	}
	adapter := fileadapter.NewAdapter("/Users/wpeng/Projects/golang/src/kit/m/policy.csv")

	e, err := casbin.NewSyncedEnforcer(m, adapter)
	if err != nil {
		return err
	}

	// load policy from file
	if err := e.LoadPolicy(); err != nil {
		return err
	}

	e.AddFunction("URIMatch", URIMatchFunctionWrapper)

	enf = e
	return nil
}

// helper: 给用户分配角色（把 user 与 role 关联）
func assignRoleToUser(userID, role string) error {
	_, err := enf.AddGroupingPolicy(userID, role)
	return err
}

// CanPerform 校验流程：
// 1) 检查订阅是否到期
// 2) 使用 casbin 判断角色/权限
// 3) 对配额类操作（例如 Screens:use）执行 TryConsume（原子）
// 注意：在分布式部署时，配额应改用 Redis 等支持并发原子操作的存储
func CanPerform(ctx context.Context, userID, obj, act string) (bool, error) {
	// 2) role resolution: if user is admin of that subscription we map to admin role.
	// we assume roles are already assigned in Casbin (assignRoleToUser)
	allowed, err := enf.Enforce(userID, obj, act)
	if err != nil {
		return false, err
	}
	if !allowed {
		return false, errors.New("forbidden by policy")
	}

	return true, nil
}

// Utility to register a user with subscription and role binding
func registerUser(userID string, pkg string, isAdmin bool, expiry time.Time) error {
	// bind roles in casbin:
	// tie user -> package role
	if _, err := enf.AddGroupingPolicy(userID, pkg); err != nil {
		return err
	}

	// if admin, also add admin role mapping to package admin capabilities
	if isAdmin {
		// we can give an admin extra role that maps to '*' permission, or map user also to backend_admin
		if _, err := enf.AddGroupingPolicy(userID, "admin"); err != nil {
			return err
		}
	}
	return nil
}

// URIMatchFunction URI决策规则函数
func URIMatchFunction(key1, key2 string) bool {
	key1 = strings.Split(key1, "?")[0]
	return casbinUtil.KeyMatch3(key1, key2)
}

// URIMatchFunctionWrapper URI决策规则函数装饰器
func URIMatchFunctionWrapper(args ...interface{}) (interface{}, error) {
	key1, _ := args[0].(string)
	key2, _ := args[1].(string)
	return URIMatchFunction(key1, key2), nil
}

func main() {
	// init
	if err := initEnforcer(); err != nil {
		log.Fatalf("initEnforcer error: %v", err)
	}
	quotaMan = NewQuotaManager()

	// 示例：创建用户和设置配额
	// 1) 后台创建订阅用户 admin 和 非admin 两种
	_ = registerUser("backend_admin_user", "admin", true, time.Now().Add(30*24*time.Hour))
	_ = registerUser("backend_normal_user", "viewer", false, time.Now().Add(30*24*time.Hour))

	// 2) 注册用户 - Starter (150 screens / month)
	_ = registerUser("starter_user", "starter", false, time.Now().Add(30*24*time.Hour))
	quotaMan.SetLimit("starter_user:Screens", 150)

	// 3) Growth (320)
	_ = registerUser("growth_user", "growth", false, time.Now().Add(30*24*time.Hour))
	quotaMan.SetLimit("growth_user:Screens", 320)

	// 4) Pro (750)
	_ = registerUser("pro_user", "pro", false, time.Now().Add(30*24*time.Hour))
	quotaMan.SetLimit("pro_user:Screens", 750)

	// 6) Trial user (3 screens)
	_ = registerUser("trial_user", "trial", false, time.Now().Add(7*24*time.Hour))
	quotaMan.SetLimit("trial_user:Screens", 3)

	// Demo: 测试不同用户的能力
	testCases := []struct {
		user string
		obj  string
		act  string
	}{
		{"backend_admin_user", "/api/v2/kyt/transactions", "POST"},
		{"growth_user", "/api/v2/kyt/engine/add", "POST"},
		{"starter_user", "/api/v2/kyt/system/update-screen-setting", "POST"},
		{"starter_user", "/api/v2/kyt/export/report/address/html", "GET"},
		{"starter_user", "/api/v2/kyt/engine/add", "POST"},
		{"growth_user", "/api/v2/kyt/system/update-screen-setting", "POST"},
		{"growth_user", "/api/v2/kyt/export/report/address/html", "GET"},
	}

	for _, tc := range testCases {
		ok, err := CanPerform(context.Background(), tc.user, tc.obj, tc.act)
		fmt.Printf("user=%s obj=%s act=%s -> ok=%v err=%v\n", tc.user, tc.obj, tc.act, ok, err)
	}

	m := casbinUtil.KeyMatch3("/api/v2/kyt/alert/1", "/api/v2/kyt/alert/{id}")
	fmt.Println(m)
}
