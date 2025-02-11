import { useEffect, useRef } from 'react'

import './App.css'
import Button from './components/button/button'
import Message from './components/message/message'
import MessagesContainer from './components/message/messagesContainer'

function App() {
  const socket = useRef<WebSocket | null>(null)
  const inputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    console.log("useEffect:", socket.current)
    if (socket.current != null) {
      return;
    }

    window.onload = function() {
      console.log("useEffect: trying to open connection")
      socket.current = new WebSocket(`ws://${window.location.host}/ws/echo`)
      socket.current.onopen = function (_) {
        console.log("OPEN");
        socket.current!.send("TEST message");
      }
      socket.current.onclose = function (_) {
        console.log("CLOSE");
        socket.current = null;
      }
      socket.current.onmessage = function (evt) {
        console.log("RESPONSE: " + evt.data);
      }
      socket.current.onerror = function (_) {
        console.log("ERROR:");
      }
    }
    
    return () => {
      if (socket.current == null) {
        return;
      }

      socket.current.close()
    }
  }, [])

  return (
    <div className="chat-container">
      <div className="chat-header">
        <h1>Text Chat #1</h1>
      </div>
      <MessagesContainer>
        <Message text="Hello! How can I help you today?"
          isUser={false}
        />
        <Message text="dunno know"
          isUser={true}
        />
      </MessagesContainer>

      <div className="input-container">
        <input ref={inputRef} type="text" className="message-input" placeholder="Type your message..." />
        <Button text="Send" onClick={() => {
          console.log("Button click!")
          if(socket.current == null) {
            return
          }

          socket.current!.send(inputRef.current!.value)
        }} />
      </div>
    </div>)
}

export default App
