package main

import (
    "fmt"
    "sync"
    "time"

    "github.com/alpacahq/alpaca-trade-api-go/alpaca"
    "github.com/alpacahq/alpaca-trade-api-go/stream"
    "github.com/loerac/ledger"
)

const notebook string = "chrisitian-loera"
var lgr ledger.Ledger
var logger *ledger.LogFile

/***
 * @brief:  Set the Alpaca URL up, and the existing ledger
 *          notebook up for reading.
 ***/
func init() {
    alpaca.SetBaseUrl("https://paper-api.alpaca.markets")
    lgr = ledger.NewLedger()
    lgr.ReadLedger(notebook + ".lgr")
    logger = ledger.NewLog("ledger.log", "listener - ")
}

/***
 * @brief:  Continually check for any updates from Alpaca Markets.
 *          Any events that occured, the `tradeHandler()` function
 *          will handle it. Only exit if error occured.
 ***/
func main() {
    ok := make(chan bool)
    var wg sync.WaitGroup

    go func() {
        if err := stream.Register(alpaca.TradeUpdates, tradeHandler); err != nil {
            ok <- false
            fmt.Println(err)
        }
    }()

    /* Sleep for a bit to not hog up CPU */
    for <-ok {
        time.Sleep(100 * time.Millisecond)
    }
    close(ok)

    wg.Wait()
}

/***
 * @brief:  Check for any event updates in any order.
 *          Events that occur are
 *              - fill
 *              - canceled
 *              - expired
 *              - replace
 *
 * @arg:    msg - The message contains the `TradeUpdate` struct
 ***/
func tradeHandler(msg interface{}) {
    tradeupdate := msg.(alpaca.TradeUpdate)
    if tradeupdate.Event == "fill" {
        acctNum := lgr.GetAcctNum(notebook + ".lgr")
        qty := tradeupdate.Order.Qty.IntPart()
        price, _ := tradeupdate.Order.FilledAvgPrice.Float64()
        price *= float64(qty)

        if  tradeupdate.Order.Side == "buy" {
            msg := fmt.Sprintf("Bought %d shares of %s", qty, tradeupdate.Order.Symbol)
            lgr.AddEntry(acctNum, tradeupdate.Order.Symbol, "Alpaca Markets", msg, -price)
        } else if tradeupdate.Order.Side == "sell" {
            msg := fmt.Sprintf("Sold %d shares of %s", qty, tradeupdate.Order.Symbol)
            lgr.AddEntry(acctNum, tradeupdate.Order.Symbol, "Alpaca Markets", msg, price)
        }
        lgr.PrintToTable(acctNum, notebook)
    } else if tradeupdate.Event == "canceled" {
        logger.Printf("Order #%s has been unfilled: %s %d shares of %s has been CANCELED\n",
            tradeupdate.Order.ID,
            tradeupdate.Order.Side,
            tradeupdate.Order.Qty.IntPart(),
            tradeupdate.Order.Symbol,
        )
    } else if tradeupdate.Event == "expired" {
        logger.Printf("Order #%s has been unfilled: %s %d shares of %s has EXPIRED - Time-in-Force(%s)\n",
            tradeupdate.Order.ID,
            tradeupdate.Order.Side,
            tradeupdate.Order.Qty.IntPart(),
            tradeupdate.Order.Symbol,
            tradeupdate.Order.TimeInForce,
        )
    } else if tradeupdate.Event == "replaced" {
        status, count, nested := "open", 1, false
        order, err := alpaca.ListOrders(&status, nil, &count, &nested)
        if err != nil {
            panic(err)
        }

        limitPrice := 0.0
        if order[0].LimitPrice != nil {
            limitPrice, _ = order[0].LimitPrice.Float64()
        }
        stopPrice := 0.0
        if order[0].StopPrice != nil {
            stopPrice, _ = order[0].StopPrice.Float64()
        }

        logger.Printf("Order #%s has been replaced by order #%s: Symbol(%s) Transaction-Type(%s) Quantity(%d) Order-Type(%s) Limit-Price(%0.3f) Stop-Price(%0.3f)\n",
                        tradeupdate.Order.ID,
                        order[0].ID,
                        order[0].Symbol,
                        order[0].Side,
                        order[0].Qty.IntPart(),
                        order[0].TimeInForce,
                        limitPrice,
                        stopPrice,
                    )
    }
}
