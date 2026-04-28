package entity

import "strings"

const cashTemplateGiftIDPrefix = "cashttpl-"

// GiftCashTemplateID é o ID determinístico do presente modelo de contribuição em dinheiro por wedding.
func GiftCashTemplateID(weddingID string) string {
	var b strings.Builder
	b.Grow(len(cashTemplateGiftIDPrefix) + len(weddingID))
	b.WriteString(cashTemplateGiftIDPrefix)
	b.WriteString(weddingID)
	return b.String()
}
