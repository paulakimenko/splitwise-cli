package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/barronlroth/splitwise-cli/internal/auth"
)

const baseURL = "https://secure.splitwise.com/api/v3.0"

// Client is a Splitwise API client.
type Client struct {
	http  *http.Client
	token string
}

// New creates a new authenticated API client.
func New() (*Client, error) {
	token, err := auth.LoadToken()
	if err != nil {
		return nil, err
	}
	return &Client{
		http: &http.Client{
			Timeout: 30 * time.Second,
		},
		token: token,
	}, nil
}

func (c *Client) get(path string, params url.Values) ([]byte, error) {
	u := baseURL + path
	if len(params) > 0 {
		u += "?" + params.Encode()
	}
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	return c.do(req)
}

func (c *Client) post(path string, params url.Values) ([]byte, error) {
	u := baseURL + path
	var body io.Reader
	if len(params) > 0 {
		body = strings.NewReader(params.Encode())
	}
	req, err := http.NewRequest("POST", u, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	if body != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	return c.do(req)
}

func (c *Client) do(req *http.Request) ([]byte, error) {
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode == 401 {
		return nil, fmt.Errorf("unauthorized — run `splitwise auth` to re-authenticate")
	}
	if resp.StatusCode == 403 {
		return nil, fmt.Errorf("forbidden — you don't have access to this resource")
	}
	if resp.StatusCode == 404 {
		return nil, fmt.Errorf("not found")
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error (HTTP %d): %s", resp.StatusCode, string(data))
	}

	return data, nil
}

// ---------- Types ----------

// User represents a Splitwise user.
type User struct {
	ID                 int64              `json:"id"`
	FirstName          string             `json:"first_name"`
	LastName           string             `json:"last_name"`
	Email              string             `json:"email"`
	RegistrationStatus string             `json:"registration_status"`
	DefaultCurrency    string             `json:"default_currency,omitempty"`
	Locale             string             `json:"locale,omitempty"`
	NotificationsCount int                `json:"notifications_count,omitempty"`
	Picture            *UserPicture       `json:"picture,omitempty"`
}

type UserPicture struct {
	Small  string `json:"small"`
	Medium string `json:"medium"`
	Large  string `json:"large"`
}

// Balance represents a currency balance.
type Balance struct {
	CurrencyCode string `json:"currency_code"`
	Amount       string `json:"amount"`
}

// Debt represents a debt between two users.
type Debt struct {
	From         int64  `json:"from"`
	To           int64  `json:"to"`
	Amount       string `json:"amount"`
	CurrencyCode string `json:"currency_code"`
}

// Group represents a Splitwise group.
type Group struct {
	ID                 int64         `json:"id"`
	Name               string        `json:"name"`
	GroupType          string        `json:"group_type"`
	UpdatedAt          string        `json:"updated_at"`
	SimplifyByDefault  bool          `json:"simplify_by_default"`
	Members            []GroupMember `json:"members"`
	OriginalDebts      []Debt        `json:"original_debts"`
	SimplifiedDebts    []Debt        `json:"simplified_debts"`
	InviteLink         string        `json:"invite_link,omitempty"`
}

// GroupMember is a user with group-specific balance info.
type GroupMember struct {
	ID        int64     `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
	Balance   []Balance `json:"balance"`
}

// Friend represents a Splitwise friend.
type Friend struct {
	ID        int64     `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
	Balance   []Balance `json:"balance"`
	Groups    []FriendGroup `json:"groups"`
}

type FriendGroup struct {
	GroupID int64     `json:"group_id"`
	Balance []Balance `json:"balance"`
}

// Expense represents a Splitwise expense.
type Expense struct {
	ID            int64          `json:"id"`
	GroupID       *int64         `json:"group_id"`
	Description   string         `json:"description"`
	Cost          string         `json:"cost"`
	CurrencyCode  string         `json:"currency_code"`
	Date          string         `json:"date"`
	Payment       bool           `json:"payment"`
	CreatedAt     string         `json:"created_at"`
	UpdatedAt     string         `json:"updated_at"`
	DeletedAt     *string        `json:"deleted_at"`
	Details       *string        `json:"details"`
	Repayments    []Repayment    `json:"repayments"`
	Users         []ExpenseShare `json:"users"`
	Category      *Category      `json:"category"`
	CreatedBy     *User          `json:"created_by"`
}

type Repayment struct {
	From   int64  `json:"from"`
	To     int64  `json:"to"`
	Amount string `json:"amount"`
}

type ExpenseShare struct {
	UserID     int64  `json:"user_id"`
	PaidShare  string `json:"paid_share"`
	OwedShare  string `json:"owed_share"`
	NetBalance string `json:"net_balance"`
	User       *User  `json:"user"`
}

type Category struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// ---------- API Methods ----------

// GetCurrentUser returns the authenticated user.
func (c *Client) GetCurrentUser() (*User, error) {
	data, err := c.get("/get_current_user", nil)
	if err != nil {
		return nil, err
	}
	var resp struct {
		User *User `json:"user"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return resp.User, nil
}

// GetGroups returns all groups for the current user.
func (c *Client) GetGroups() ([]Group, error) {
	data, err := c.get("/get_groups", nil)
	if err != nil {
		return nil, err
	}
	var resp struct {
		Groups []Group `json:"groups"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return resp.Groups, nil
}

// GetGroup returns a single group by ID.
func (c *Client) GetGroup(id int64) (*Group, error) {
	data, err := c.get(fmt.Sprintf("/get_group/%d", id), nil)
	if err != nil {
		return nil, err
	}
	var resp struct {
		Group *Group `json:"group"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return resp.Group, nil
}

// GetFriends returns all friends for the current user.
func (c *Client) GetFriends() ([]Friend, error) {
	data, err := c.get("/get_friends", nil)
	if err != nil {
		return nil, err
	}
	var resp struct {
		Friends []Friend `json:"friends"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return resp.Friends, nil
}

// GetExpensesParams holds query parameters for listing expenses.
type GetExpensesParams struct {
	GroupID     int64
	FriendID   int64
	DatedAfter string
	DatedBefore string
	Limit      int
	Offset     int
}

// GetExpenses returns expenses matching the given criteria.
func (c *Client) GetExpenses(p GetExpensesParams) ([]Expense, error) {
	params := url.Values{}
	if p.GroupID > 0 {
		params.Set("group_id", fmt.Sprintf("%d", p.GroupID))
	}
	if p.FriendID > 0 {
		params.Set("friend_id", fmt.Sprintf("%d", p.FriendID))
	}
	if p.DatedAfter != "" {
		params.Set("dated_after", p.DatedAfter)
	}
	if p.DatedBefore != "" {
		params.Set("dated_before", p.DatedBefore)
	}
	if p.Limit > 0 {
		params.Set("limit", fmt.Sprintf("%d", p.Limit))
	}
	if p.Offset > 0 {
		params.Set("offset", fmt.Sprintf("%d", p.Offset))
	}
	data, err := c.get("/get_expenses", params)
	if err != nil {
		return nil, err
	}
	var resp struct {
		Expenses []Expense `json:"expenses"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return resp.Expenses, nil
}

// CreateExpenseParams holds parameters for creating an expense.
type CreateExpenseParams struct {
	Description  string
	Cost         string
	CurrencyCode string
	GroupID      int64
	SplitEqually bool
	Date         string
	// For by-shares split: user_id -> {paid_share, owed_share}
	Shares []ShareParam
}

type ShareParam struct {
	UserID    int64
	PaidShare string
	OwedShare string
}

// CreateExpense creates a new expense.
func (c *Client) CreateExpense(p CreateExpenseParams) (*Expense, error) {
	params := url.Values{}
	params.Set("description", p.Description)
	params.Set("cost", p.Cost)
	if p.CurrencyCode != "" {
		params.Set("currency_code", p.CurrencyCode)
	}
	if p.Date != "" {
		params.Set("date", p.Date)
	}

	if p.SplitEqually {
		params.Set("group_id", fmt.Sprintf("%d", p.GroupID))
		params.Set("split_equally", "true")
	} else if len(p.Shares) > 0 {
		if p.GroupID > 0 {
			params.Set("group_id", fmt.Sprintf("%d", p.GroupID))
		} else {
			params.Set("group_id", "0")
		}
		for i, s := range p.Shares {
			prefix := fmt.Sprintf("users__%d__", i)
			params.Set(prefix+"user_id", fmt.Sprintf("%d", s.UserID))
			params.Set(prefix+"paid_share", s.PaidShare)
			params.Set(prefix+"owed_share", s.OwedShare)
		}
	}

	data, err := c.post("/create_expense", params)
	if err != nil {
		return nil, err
	}

	var resp struct {
		Expenses []Expense              `json:"expenses"`
		Errors   map[string]interface{} `json:"errors"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	if len(resp.Errors) > 0 {
		errData, _ := json.Marshal(resp.Errors)
		return nil, fmt.Errorf("expense creation failed: %s", string(errData))
	}
	if len(resp.Expenses) == 0 {
		return nil, fmt.Errorf("no expense returned")
	}
	return &resp.Expenses[0], nil
}

// DeleteExpense deletes an expense by ID.
func (c *Client) DeleteExpense(id int64) error {
	data, err := c.post(fmt.Sprintf("/delete_expense/%d", id), nil)
	if err != nil {
		return err
	}
	var resp struct {
		Success bool                   `json:"success"`
		Errors  map[string]interface{} `json:"errors"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}
	if !resp.Success {
		errData, _ := json.Marshal(resp.Errors)
		return fmt.Errorf("delete failed: %s", string(errData))
	}
	return nil
}

// CreatePayment records a settlement (payment) between users.
func (c *Client) CreatePayment(p CreateExpenseParams) (*Expense, error) {
	params := url.Values{}
	params.Set("description", "Payment")
	params.Set("cost", p.Cost)
	params.Set("payment", "true")
	if p.CurrencyCode != "" {
		params.Set("currency_code", p.CurrencyCode)
	}
	if p.GroupID > 0 {
		params.Set("group_id", fmt.Sprintf("%d", p.GroupID))
	} else {
		params.Set("group_id", "0")
	}
	if p.Date != "" {
		params.Set("date", p.Date)
	}
	for i, s := range p.Shares {
		prefix := fmt.Sprintf("users__%d__", i)
		params.Set(prefix+"user_id", fmt.Sprintf("%d", s.UserID))
		params.Set(prefix+"paid_share", s.PaidShare)
		params.Set(prefix+"owed_share", s.OwedShare)
	}

	data, err := c.post("/create_expense", params)
	if err != nil {
		return nil, err
	}

	var resp struct {
		Expenses []Expense              `json:"expenses"`
		Errors   map[string]interface{} `json:"errors"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	if len(resp.Errors) > 0 {
		errData, _ := json.Marshal(resp.Errors)
		return nil, fmt.Errorf("settlement failed: %s", string(errData))
	}
	if len(resp.Expenses) == 0 {
		return nil, fmt.Errorf("no expense returned")
	}
	return &resp.Expenses[0], nil
}

// ResolveGroupByName finds a group ID by name (case-insensitive partial match).
func (c *Client) ResolveGroupByName(name string) (*Group, error) {
	groups, err := c.GetGroups()
	if err != nil {
		return nil, err
	}
	lower := strings.ToLower(name)
	for i := range groups {
		if strings.ToLower(groups[i].Name) == lower {
			return &groups[i], nil
		}
	}
	// Partial match fallback.
	for i := range groups {
		if strings.Contains(strings.ToLower(groups[i].Name), lower) {
			return &groups[i], nil
		}
	}
	return nil, fmt.Errorf("group not found: %s", name)
}

// ResolveFriendByName finds a friend by name (case-insensitive).
func (c *Client) ResolveFriendByName(name string) (*Friend, error) {
	friends, err := c.GetFriends()
	if err != nil {
		return nil, err
	}
	lower := strings.ToLower(name)
	for i := range friends {
		fullName := strings.ToLower(friends[i].FirstName + " " + friends[i].LastName)
		if fullName == lower || strings.ToLower(friends[i].FirstName) == lower {
			return &friends[i], nil
		}
	}
	// Partial match.
	for i := range friends {
		fullName := strings.ToLower(friends[i].FirstName + " " + friends[i].LastName)
		if strings.Contains(fullName, lower) {
			return &friends[i], nil
		}
	}
	return nil, fmt.Errorf("friend not found: %s", name)
}
