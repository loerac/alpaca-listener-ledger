# Alpaca-Ledger
Using the [alpaca.market](https://github.com/alpacahq/alpaca-trade-api-go) API, listen for any event updates. Depending on the events, it'll log the action that occured.
This repository also uses the [ledger](https://github.com/loerac/ledger/) API to add the transaction entries to the ledger notebook.

#### Note:
You will need a pre-existing ledger notebook to do this. In order to get a ledger notebook, create one by running the ledger API [example](https://github.com/loerac/ledger/tree/trunk/example) and then update the notebook filename with the your notebook filename.
