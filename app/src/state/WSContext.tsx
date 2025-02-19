import { createContext, useRef } from "react"

import { useWebSocket, WSMessage } from "../hooks/useWebSocket"

type onMessageFunc = (message: WSMessage) => void
type onErrorFunc = () => void

interface IWSCallbacks {
    onMessage: Array<onMessageFunc>
    onError: Array<onErrorFunc>
}

interface IWSContext {
    addOnMessageCallback: (callback: onMessageFunc) => void
    removeOnMessageCallback: (callback: onMessageFunc) => void
    addOnErrorCallback: (callback: onErrorFunc) => void
    removeOnErrorCallback: (callback: onErrorFunc) => void
    send: (message: string) => void
}

export const WSContext = createContext<IWSContext>({
    addOnErrorCallback: () => { },
    addOnMessageCallback: () => { },
    removeOnErrorCallback: () => { },
    removeOnMessageCallback: () => { },
    send: () => { },
})

export function WSProvider(props: React.PropsWithChildren) {
    const wscallbacks = useRef<IWSCallbacks>({
        onError: [],
        onMessage: [],
    })

    const { send } = useWebSocket({
        path: "/ws/completions",
        onMessage(message) {
            wscallbacks.current.onMessage.forEach(callback => {
                callback(message)
            })
        },
        onError() {
            wscallbacks.current.onError.forEach(callback => {
                callback()
            })
        },
    })

    function addOnMessageCallback(callback: (message: WSMessage) => void) {
        if (!wscallbacks.current.onMessage.includes(callback)){
            wscallbacks.current.onMessage.push(callback)
        }
    }

    function removeOnMessageCallback(callback: (message: WSMessage) => void) {
        const index = wscallbacks.current.onMessage.findIndex(reference => reference === callback)
        if (index != -1) {
            wscallbacks.current.onMessage = wscallbacks.current.onMessage.splice(index, 1)
        }
    }

    function addOnErrorCallback(callback: () => void) {
        if (!wscallbacks.current.onError.includes(callback))
            wscallbacks.current.onError.push(callback)
    }

    function removeOnErrorCallback(callback: () => void) {
        const index = wscallbacks.current.onError.findIndex(reference => reference === callback)
        if (index != -1) {
            wscallbacks.current.onError = wscallbacks.current.onError.splice(index, 1)
        }
    }

    return (<WSContext.Provider value={{
        addOnMessageCallback,
        removeOnMessageCallback,
        addOnErrorCallback,
        removeOnErrorCallback,
        send
    }}>
        {props.children}
    </WSContext.Provider>)
}