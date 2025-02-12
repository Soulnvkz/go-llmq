import { useRef, useState } from 'react'

import './App.css'
import Button from './components/button/button'
import Message from './components/message/message'
import MessagesContainer from './components/message/messagesContainer'
import useWebSocket from './hooks/useWebSocket'

interface Message {
  id: number;
  text: string;
  isUser: boolean;
}

function App() {
  const { send } = useWebSocket({
    path: "/ws/echo",
    onMessage: (message: string) => {
      console.log("onMessage", message);
      if (message === "<start>") {
        currentRef.current!.innerHTML = "<span></span>"
        currentRef.current!.style.display = "block"
        return
      }
      if (message === "<end>") {
        index.current = index.current + 1;
        const text = currentRef.current!.innerHTML.replace("<span>", "").replace("</span>", "")
        currentRef.current!.innerHTML = "<span></span>"
        currentRef.current!.style.display = "none"
        setMessages(prev => [...prev, {
          id: index.current,
          isUser: false,
          text: text
        }]);
        
        return
      }
      if(currentRef.current) {
        currentRef.current.innerHTML = "<span>" + currentRef.current.innerHTML.replace("<span>", "").replace("</span>", "") + message + "</span>"
      }
    }
  })

  const index = useRef(0);
  const inputRef = useRef<HTMLInputElement>(null);
  const [messages, setMessages] = useState<Message[]>([]);
  const currentRef = useRef<HTMLDivElement>(null);

  return (
    <div className="chat-container">
      <div className="chat-header">
        <h1>Text Chat #1</h1>
      </div>
      <MessagesContainer>
        {messages.map(m => <Message key={m.id} text={m.text}
          isUser={m.isUser}
        />)}
        <Message last ref={currentRef} isUser={false} />
      </MessagesContainer>

      <div className="input-container">
        <input ref={inputRef} type="text" className="message-input" placeholder="Type your message..." />
        <Button text="Send" onClick={() => {
          const message = inputRef.current!.value
          if (message.length === 0) {
            return;
          }

          index.current = index.current + 1;
          inputRef.current!.value = ""
          setMessages(prev => [...prev, {
            id: index.current,
            isUser: true,
            text: message
          }]);

          send(message);
        }} />
      </div>
    </div>)
}

export default App
