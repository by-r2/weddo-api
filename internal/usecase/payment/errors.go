package payment

import "errors"

var (
	ErrCheckoutEmptyItems           = errors.New("informe pelo menos um item")
	ErrCheckoutDuplicateGiftLine    = errors.New("presente repetido no pedido — remova linhas duplicadas")
	ErrCheckoutInvalidCatalogExtras = errors.New("itens do catálogo não aceitam valor ou texto personalizado no item")
	ErrCheckoutCashAmountMissing    = errors.New("informe o valor para a contribuição em dinheiro")
)
