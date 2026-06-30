package cache

import "fmt"

func (c *Client) AccountTokenKey(accountID uint) string {
	prefix := ""
	if c != nil {
		prefix = c.keyPrefix + ":"
	}
	return prefix + fmt.Sprintf("account:token:%d", accountID)
}

func (c *Client) VideoDetailKey(videoID uint) string {
	prefix := ""
	if c != nil {
		prefix = c.keyPrefix + ":"
	}
	return prefix + fmt.Sprintf("video:detail:%d", videoID)
}

func (c *Client) ProfileKey(accountID uint) string {
	prefix := ""
	if c != nil {
		prefix = c.keyPrefix + ":"
	}
	return prefix + fmt.Sprintf("profile:%d", accountID)
}

func (c *Client) FeedLatestKey() string {
	prefix := ""
	if c != nil {
		prefix = c.keyPrefix + ":"
	}
	return prefix + "feed:latest"
}
