workspace valkyrie {
    model {
        casinoOperator = element "Casino operator" "Running one or more instances of Valkyrie"
        gameProvider = element "Game Provider"
        valkyrie = softwareSystem "Valkyrie system" "Valkyrie open source igaming aggregator"{
            valk = container "Valkyrie"{
                pam = component "PAM Wallet" "Implementation of PamClient. Making calls to operator, updating account balance etc."
                operatorServer = component "Operator server" "Exposing port used by operator to make requests toward Valkyrie and Game providers"
                providerServer = component "Provider Server" "Exposing port used by game provider to make requests toward Valkyrie"
                group Provider {
                    operatorRouter = component "Operator Router" "Implementing 'operator_api.yml' oapi specification"
                    providerClient = component "Provider Client" "Handle authentication toward game provider"
                    providerRouter = component "Provider Router" "Implementing the specific provider's wallet api. Maps the requests to conform to Valkyrie PAM api"

                }
            }
            pamVplugin = container "PAM Plugin" "OPTIONAL. PAM plugin running as a separate process"
        }
        casinoOperator -> operatorServer "Makes api calls to"
        operatorServer -> operatorRouter "Routes to specific provider"
        operatorRouter -> providerClient "calls"
        providerClient -> gameProvider "Makes api calls to"
        providerServer -> providerRouter "Routes to specific provider"
        providerRouter -> pam "Calls"
        pam -> casinoOperator "Makes api calls to"
        pam -> pamVplugin "Calls" 
        pamVplugin -> casinoOperator "Makes api calls to"
        gameProvider -> providerServer "Makes api calls to"

    }
    views {
        theme default
        systemLandscape {
            !script groovy {
                workspace.views.createDefaultViews()
                workspace.views.views.each { it.disableAutomaticLayout() }
            }
        }
    }
}