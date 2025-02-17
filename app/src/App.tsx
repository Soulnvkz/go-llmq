import { useRef, useState } from 'react'

import './App.css'
import Button from './components/button/button'
import Message from './components/message/message'
import MessagesContainer from './components/message/messagesContainer'
import { CancelMessage, CompletitionsEnd, CompletitionsMessage, CompletitionsNext, CompletitionsQueue, CompletitionsStart, useWebSocket, WSMessage } from './hooks/useWebSocket'



interface Message {
  id: number;
  text: string;
  isUser: boolean;
}

function App() {
  const { send } = useWebSocket({
    path: "/ws/completions",
    onMessage: (message: WSMessage) => {
      switch (message.message_type) {
        case CompletitionsQueue:
          currentRef.current!.innerHTML = "<span>queue...</span>"
          currentRef.current!.style.display = "block"
          break;
        case CompletitionsStart:
          currentRef.current!.innerHTML = "<span></span>"
          currentRef.current!.style.display = "block"
          break;
        case CompletitionsNext:
          if (currentRef.current) {
            currentRef.current.innerHTML = "<span>" + currentRef.current.innerHTML.replace("<span>", "").replace("</span>", "") + message.content! + "</span>"
          }
          break;
        case CompletitionsEnd:
          index.current = index.current + 1;
          const text = currentRef.current!.innerHTML.replace("<span>", "").replace("</span>", "")
          currentRef.current!.innerHTML = "<span></span>"
          currentRef.current!.style.display = "none"
          setMessages(prev => [...prev, {
            id: index.current,
            isUser: false,
            text: text
          }]);
          break;
        default:
          console.info("unsupported message type", message.message_type);
          break;
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
        <Button text="Cancel" onClick={() => {
          send(JSON.stringify({
            message_type: CancelMessage,
          }));
        }} />
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

          send(JSON.stringify({
            message_type: CompletitionsMessage,
            content: message
          }));
        }} />
      </div>
    </div>)
}

export default App
