package auth

import "time"

type User struct {
	ID                    int64     `json:"id"`
	Username              string    `json:"username"`
	Nickname              string    `json:"nickname"`
	RealName              string    `json:"realName"`
	Phone                 string    `json:"phone"`
	Email                 string    `json:"email"`
	Status                int       `json:"status"`
	DeptID                int64     `json:"deptId"`
	RoleKeys              []string  `json:"roleKeys"`
	RoleNames             []string  `json:"roleNames"`
	MenuKeys              []string  `json:"menuKeys"`
	PasswordResetRequired int       `json:"passwordResetRequired"`
	CreatedAt             time.Time `json:"createdAt"`
	UpdatedAt             time.Time `json:"updatedAt"`
}

type Department struct {
	ID        int64  `json:"id"`
	ParentID  int64  `json:"parentId"`
	Ancestors string `json:"ancestors"`
	Name      string `json:"name"`
	Leader    string `json:"leader"`
	Phone     string `json:"phone"`
	Email     string `json:"email"`
	Sort      int    `json:"sort"`
	Status    int    `json:"status"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

type UserRecord struct {
	ID                    int64    `json:"id"`
	DeptID                int64    `json:"deptId"`
	DeptName              string   `json:"deptName"`
	Username              string   `json:"username"`
	Nickname              string   `json:"nickname"`
	RealName              string   `json:"realName"`
	Phone                 string   `json:"phone"`
	Email                 string   `json:"email"`
	Status                int      `json:"status"`
	LastLoginIP           string   `json:"lastLoginIp"`
	LastLoginAt           string   `json:"lastLoginAt"`
	RoleIDs               []int64  `json:"roleIds"`
	RoleNames             []string `json:"roleNames"`
	PasswordResetRequired int      `json:"passwordResetRequired"`
	CreatedAt             string   `json:"createdAt"`
	UpdatedAt             string   `json:"updatedAt"`
}

type Role struct {
	ID          int64    `json:"id"`
	Name        string   `json:"name"`
	Key         string   `json:"key"`
	DataScope   string   `json:"dataScope"`
	Sort        int      `json:"sort"`
	Status      int      `json:"status"`
	Description string   `json:"description"`
	MenuKeys    []string `json:"menuKeys"`
	CreatedAt   string   `json:"createdAt"`
	UpdatedAt   string   `json:"updatedAt"`
}

type LoginLog struct {
	ID          int64  `json:"id"`
	UserID      int64  `json:"userId"`
	Username    string `json:"username"`
	LoginIP     string `json:"loginIp"`
	UserAgent   string `json:"userAgent"`
	LoginStatus int    `json:"loginStatus"`
	Message     string `json:"message"`
	CreatedAt   string `json:"createdAt"`
}

type Menu struct {
	ID        int64  `json:"id"`
	ParentID  int64  `json:"parentId"`
	Name      string `json:"name"`
	MenuType  string `json:"menuType"`
	Path      string `json:"path"`
	Component string `json:"component"`
	RouteName string `json:"routeName"`
	Perms     string `json:"perms"`
	Icon      string `json:"icon"`
	Visible   int    `json:"visible"`
	Status    int    `json:"status"`
	Sort      int    `json:"sort"`
}
