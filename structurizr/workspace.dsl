workspace valkyrie {
    model {
        user = person "Player" "User with a Gaming Operator, eg an online casino"
        game = element "Game" "Game UI provided by Game Provider. Bets etc communicated toward Game provider backend."
        casinoOperator = element "Gaming operator" "Running one or more instances of Valkyrie"
        gameProvider = element "Game Provider" "Developing games and hosting gaming servers"
        valkyrie = softwareSystem "Valkyrie system" "Valkyrie open source igaming aggregator"{
            valk = container "Valkyrie"{
                technology "Go"
                pam = component "PAM Client" "Implementation of PamClient. Making calls to operator, updating account balance etc."
                operatorServer = component "Operator server" "Exposing port used by operator to make requests toward Valkyrie and Game providers" "Go fiber"
                providerServer = component "Provider Server" "Exposing port used by game provider to make requests toward Valkyrie" "Go fiber"
                group "Provider module" {
                    operatorRouter = component "Operator Router" "Implementing 'operator_api.yml' oapi specification" "Go fiber"
                    providerClient = component "Provider Client" "Handling authentication towards game provider"
                    providerMapper = component "Provider Mapper" "Mapping the provider wallet request to Valkyrie standard requests"
                    providerRouter = component "Provider Router" "Implementing the specific provider's wallet api." "Go fiber"

                }
            }
            pamVplugin = container "PAM Plugin" "OPTIONAL. PAM plugin running as a separate process. Needs to implement specified vplugin interface."
        }

        user -> casinoOperator "Finds games to play"
        casinoOperator -> user "Gives url to Game"
        user -> game "Playing"
        game -> gameProvider "Makes api calls to"
        casinoOperator -> operatorServer "Makes api calls to" "HTTPS"
        operatorServer -> operatorRouter "Routes to specific provider"
        operatorRouter -> providerClient "calls"
        providerClient -> gameProvider "Makes api calls to" "HTTPS"
        providerServer -> providerRouter "Routes to specific provider"
        providerRouter -> providerMapper "Calls"
        providerMapper -> pam "Calls"
        pam -> casinoOperator "Makes wallet api calls" "HTTPS/JSON"
        pam -> pamVplugin "Calls" "RPC/GRPC"
        pamVplugin -> casinoOperator "Makes wallet api calls to"
        gameProvider -> providerServer "Makes wallet api calls to" "HTTPS"

    }
    views {
        theme default
        !script groovy {
            workspace.views.views.each { it.disableAutomaticLayout() }
        }
        systemLandscape OnlineGaming "Online gaming with Valkyrie"{
            include *
        }
        systemContext valkyrie "Valkyrie-System"{
            include *
        }

        container valkyrie "Valkyrie-Containers"{
            include *

        }
        component valk "Valkyrie-Components" {
            include *
        }
        styles {
            element "Group" {
                color #444444
                fontSize 28
            }
        }
    }
}