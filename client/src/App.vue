<script setup lang="ts">
import { io } from "socket.io-client";
import { ref, onMounted, onUnmounted } from "vue";

// Интерфейс для превью ссылки
interface LinkPreview {
  url: string;
  title: string;
  description: string;
  image: string;
  siteName: string;
  messageText: string; // Оригинальный текст сообщения
}

// Интерфейс для превью файла
interface FilePreview {
  url: string;
  fileName: string;
  fileType: string;
  viewerUrl: string;
  messageText: string;
}

// Интерфейс для реакции
interface Reaction {
  emoji: string;
  count: number;
  users: string[];
}

// Интерфейс для комнаты
interface Room {
  id: string;
  name: string;
  description: string;
  createdAt: number;
  userCount: number;
}

// Интерфейс для сообщения
interface Message {
  id: string;
  type: "text" | "preview" | "file" | "image" | "video";
  content: string | LinkPreview | FilePreview | any;
  reactions?: Record<string, Reaction>;
  timestamp: number;
}

const messages = ref<Message[]>([]);
const message = ref("");
const isModalOpen = ref(false);
const modalFileUrl = ref("");
const isDragOver = ref(false);
const isUploading = ref(false);
const showEmojiPicker = ref(false);
const showMoreEmoji = ref(false);
const selectedMessageId = ref("");
const currentUserId = ref("user_" + Math.random().toString(36).substr(2, 9));
const emojiPickerMode = ref<"text" | "reaction">("text");

// Состояние комнат
const rooms = ref<Room[]>([]);
const currentRoom = ref<Room | null>(null);
const showRoomModal = ref(false);
const newRoomName = ref("");
const newRoomDescription = ref("");

const socket = io(
  // "https://api.logistics.sriss.uz/",
  // "http://localhost:3300/", 
  "http://localhost:8080/", 
  {
    transports: ["websocket"],
    auth: {
      token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyIjp7ImlkIjoxMSwicm9sZUlkIjoxLCJvcmdJZCI6MSwic29hdG9JZCI6MTcsImlzQWRtaW4iOnRydWV9LCJleHAiOjE3NTcwMjUxMzIsIm5iZiI6MTc1Njk4OTEzMiwiaWF0IjoxNzU2OTg5MTMyLCJqdGkiOiIxMSJ9.BJRuINX0xXowW6ToGPac44DomuimIcmCnucc6mUuNrc",
    },
  }
);

socket.on("connect", () => {
  console.log("🟢 Подключено к серверу");
  socket.emit("connected", { organizationId: 1 });
  console.log("Emitted 'connected' with organizationId: 1");
});

// Обработчики для комнат
socket.on("rooms list", function (roomsList: Room[]) {
  console.log("📋 Получен список комнат:", roomsList);
  rooms.value = roomsList;
});

socket.on("room joined", function (room: Room) {
  console.log("🏠 Присоединились к комнате:", room);
  currentRoom.value = room;
  messages.value = []; // Очищаем сообщения при смене комнаты
});

// Обработчик истории сообщений комнаты
socket.on("room history", function (history: any[]) {
  console.log("История: ===========>", history);

  const historyMessages: Message[] = [];

  history.forEach((msg) => {
    historyMessages.push({
      id: msg.id || "msg_" + msg.timestamp + "_" + Math.random().toString(36).substr(2, 9),
      type: msg.type,
      content: msg.content,
      reactions: {},
      timestamp: msg.timestamp,
    });
  });


  messages.value = historyMessages;

  setTimeout(() => { window.scrollTo(0, document.body.scrollHeight) }, 100);
});

socket.on("room created", function (room: Room) {
  console.log("🏠 Создана новая комната:", room);
  showRoomModal.value = false;
  newRoomName.value = "";
  newRoomDescription.value = "";
});

socket.on("chat message", function (msg: string) {
  console.log("📩 Новое сообщение:", msg);
  messages.value.push({
    id: "msg_" + Date.now() + "_" + Math.random().toString(36).substr(2, 9),
    type: "text",
    content: msg,
    reactions: {},
    timestamp: Date.now(),
  });
  window.scrollTo(0, document.body.scrollHeight);
});

// Обработчик для превью ссылок
socket.on("link preview", function (preview: LinkPreview) {
  console.log("🔗 Получено превью ссылки:", preview);
  const messageIndex = messages.value.findIndex(
    (msg) => msg.type === "text" && msg.content === preview.messageText
  );

  console.log("🔗 Найдено сообщение для замены на превью:", messageIndex, preview);

  if (messageIndex !== -1) {
    // Заменяем сообщение на превью
    messages.value[messageIndex] = {
      ...messages.value[messageIndex],
      type: "preview",
      content: preview,
    };
  } else {
    // Если не нашли, просто добавляем превью
    messages.value.push({
      id: "msg_" + Date.now() + "_" + Math.random().toString(36).substr(2, 9),
      type: "preview",
      content: preview,
      reactions: {},
      timestamp: Date.now(),
    });
  }

  window.scrollTo(0, document.body.scrollHeight);
});

// Обработчик для превью файлов
socket.on("file preview", function (filePreview: FilePreview) {
  // Находим сообщение с оригинальным текстом и заменяем его на превью файла
  const messageIndex = messages.value.findIndex(
    (msg) => msg.type === "text" && msg.content === filePreview.messageText
  );

  console.log("📎 Найдено сообщение для замены на превью файла:", messageIndex, filePreview);

  if (messageIndex !== -1) {
    // Заменяем сообщение на превью файла
    messages.value[messageIndex] = {
      ...messages.value[messageIndex],
      type: "file",
      content: filePreview,
    };
  } else {
    // Если не нашли, просто добавляем превью файла
    messages.value.push({
      id: "msg_" + Date.now() + "_" + Math.random().toString(36).substr(2, 9),
      type: "file",
      content: filePreview,
      reactions: {},
      timestamp: Date.now(),
    });
  }

  window.scrollTo(0, document.body.scrollHeight);
});

// Обработчик для изображений
socket.on("image message", function (imageData: any) {
  messages.value.push({
    id: "msg_" + Date.now() + "_" + Math.random().toString(36).substr(2, 9),
    type: "image",
    content: imageData,
    reactions: {},
    timestamp: Date.now(),
  });
  window.scrollTo(0, document.body.scrollHeight);
});

// Обработчик для видео
socket.on("video message", function (videoData: any) {
  messages.value.push({
    id: "msg_" + Date.now() + "_" + Math.random().toString(36).substr(2, 9),
    type: "video",
    content: videoData,
    reactions: {},
    timestamp: Date.now(),
  });
  window.scrollTo(0, document.body.scrollHeight);
});

// Обработчик для реакций
socket.on(
  "reaction",
  function (data: {
    messageId: string;
    emoji: string;
    userId: string;
    action: "add" | "remove";
  }) {
    const messageIndex = messages.value.findIndex(
      (msg) => msg.id === data.messageId
    );
    if (messageIndex !== -1) {
      const message = messages.value[messageIndex];
      if (!message.reactions) {
        message.reactions = {};
      }

      if (data.action === "add") {
        if (!message.reactions[data.emoji]) {
          message.reactions[data.emoji] = {
            emoji: data.emoji,
            count: 0,
            users: [],
          };
        }

        if (!message.reactions[data.emoji].users.includes(data.userId)) {
          message.reactions[data.emoji].users.push(data.userId);
          message.reactions[data.emoji].count++;
        }
      } else {
        if (message.reactions[data.emoji]) {
          const userIndex = message.reactions[data.emoji].users.indexOf(
            data.userId
          );
          if (userIndex > -1) {
            message.reactions[data.emoji].users.splice(userIndex, 1);
            message.reactions[data.emoji].count--;

            if (message.reactions[data.emoji].count === 0) {
              delete message.reactions[data.emoji];
            }
          }
        }
      }
    }
  }
);

function openFileModal(viewerUrl: string, fileType?: string) {
  console.log("👁️ Открываем модальное окно для просмотра файла:", viewerUrl, fileType);

  if (fileType === "pdf") {
    modalFileUrl.value = viewerUrl;
    isModalOpen.value = true;
  } else if (fileType === "image") {
    modalFileUrl.value = viewerUrl;
    isModalOpen.value = true;
  } else if (fileType === "video") {
    modalFileUrl.value = viewerUrl;
    isModalOpen.value = true;
  } else if (
    fileType === "document" ||
    fileType === "spreadsheet" ||
    fileType === "presentation"
  ) {
    modalFileUrl.value = viewerUrl;
    isModalOpen.value = true;
  } else if (fileType === "file") {
    modalFileUrl.value = viewerUrl;
    isModalOpen.value = true;
  } else {
    window.open(viewerUrl, "_blank");
    return;
  }
}

// Функция для закрытия модального окна
function closeModal() {
  isModalOpen.value = false;
  modalFileUrl.value = "";
}

// Функция для загрузки файла
async function uploadFile(file: File) {
  if (isUploading.value) return;

  isUploading.value = true;

  const formData = new FormData();
  formData.append("file", file);

  try {
    const response = await fetch("http://localhost:3300/upload", {
      method: "POST",
      body: formData,
    });

    if (!response.ok) {
      throw new Error("Ошибка загрузки файла");
    }

    const uploadedFile = await response.json();

    // Отправляем информацию о файле через Socket.IO
    socket.emit("file uploaded", uploadedFile);
  } catch (error) {
    console.error("❌ Ошибка загрузки:", error);
    alert("❌ Ошибка загрузки файла");
  } finally {
    isUploading.value = false;
  }
}

// Функция для загрузки множественных файлов
async function uploadMultipleFiles(files: FileList) {
  for (let i = 0; i < files.length; i++) {
    await uploadFile(files[i]);
    // Небольшая задержка между загрузками
    await new Promise((resolve) => setTimeout(resolve, 300));
  }
}

// Обработчики drag & drop
function onDragOver(e: DragEvent) {
  e.preventDefault();
  e.stopPropagation();
  isDragOver.value = true;
}

function onDragLeave(e: DragEvent) {
  e.preventDefault();
  e.stopPropagation();
  // Проверяем, что мы действительно покинули зону
  if (!e.currentTarget?.contains(e.relatedTarget as Node)) {
    isDragOver.value = false;
  }
}

function onDrop(e: DragEvent) {
  e.preventDefault();
  e.stopPropagation();
  isDragOver.value = false;

  const files = e.dataTransfer?.files;
  if (files && files.length > 0) {
    uploadMultipleFiles(files);
  }
}

// Глобальные обработчики для drag & drop
function onGlobalDragOver(e: DragEvent) {
  e.preventDefault();
  isDragOver.value = true;
}

function onGlobalDragLeave(e: DragEvent) {
  if (e.clientX === 0 && e.clientY === 0) {
    isDragOver.value = false;
  }
}

function onGlobalDrop(e: DragEvent) {
  e.preventDefault();
  isDragOver.value = false;
}

// Обработчик выбора файла через input
function onFileSelect(e: Event) {
  const input = e.target as HTMLInputElement;
  const files = input.files;
  if (files && files.length > 0) {
    uploadMultipleFiles(files);
  }
  input.value = ""; // Сбрасываем input
}

// Функция для открытия в Office Online
function openInOfficeOnline() {
  const fileUrl = modalFileUrl.value.includes("officeapps")
    ? modalFileUrl.value.split("src=")[1]
    : modalFileUrl.value;
  const officeUrl = `https://view.officeapps.live.com/op/embed.aspx?src=${encodeURIComponent(
    fileUrl
  )}`;
  window.open(officeUrl, "_blank");
}

// Обработчик ошибки iframe
function onIframeError() {
  console.log("❌ Ошибка загрузки в iframe");
}

// Обработчик copy-paste файлов
function onPaste(e: ClipboardEvent) {
  const items = e.clipboardData?.items;
  if (!items) return;

  for (let i = 0; i < items.length; i++) {
    const item = items[i];

    // Проверяем, является ли элемент файлом
    if (item.kind === "file") {
      const file = item.getAsFile();
      if (file) {
        uploadFile(file);
      }
    }
  }
}

function submit() {
  if (!message.value) return;
  socket.emit("chat message", {roomId: "fbe8168d-237e-40c1-bc7e-31be03ba8b93", content: message.value},);
  message.value = "";
}

// Функции для работы с реакциями
function showEmojiPickerForMessage(messageId: string) {
  selectedMessageId.value = messageId;
  emojiPickerMode.value = "reaction";
  showEmojiPicker.value = true;
}

function showEmojiPickerForText() {
  emojiPickerMode.value = "text";
  showEmojiPicker.value = true;
}

function addEmojiToText(emoji: string) {
  message.value += emoji;

  showEmojiPicker.value = false;
  showMoreEmoji.value = false;

  const inputElement = document.getElementById("input") as HTMLInputElement;
  if (inputElement) {
    inputElement.focus();
  }
}

function addReaction(messageId: string, emoji: string) {
  const message = messages.value.find((msg) => msg.id === messageId);
  if (!message) return;

  if (!message.reactions) {
    message.reactions = {};
  }

  const hasReaction = message.reactions[emoji]?.users.includes(
    currentUserId.value
  );

  if (hasReaction) {
    socket.emit("reaction", {
      messageId,
      emoji,
      userId: currentUserId.value,
      action: "remove",
    });
  } else {
    socket.emit("reaction", {
      messageId,
      emoji,
      userId: currentUserId.value,
      action: "add",
    });
  }

  if (emojiPickerMode.value === "reaction") {
    showEmojiPicker.value = false;
    showMoreEmoji.value = false;
  }
}

function getReactionClass(messageId: string, emoji: string): string {
  const message = messages.value.find((msg) => msg.id === messageId);
  const hasReaction = message?.reactions?.[emoji]?.users.includes(
    currentUserId.value
  );
  return hasReaction ? "reaction-active" : "reaction-inactive";
}

function closeEmojiPicker() {
  showEmojiPicker.value = false;
  showMoreEmoji.value = false;
}

function joinRoom(room: Room) {
  console.log("🏠 Присоединяемся к комнате:", room.name);
  console.log("Room ID:", room.id);
  socket.emit("join room", { id: room.id });
}

function createRoom() {
  if (!newRoomName.value.trim()) return;

  socket.emit("create room", {
    name: newRoomName.value.trim(),
    description: newRoomDescription.value.trim(),
  });
}

function showCreateRoomModal() {
  showRoomModal.value = true;
}

function closeRoomModal() {
  showRoomModal.value = false;
  newRoomName.value = "";
  newRoomDescription.value = "";
}

onMounted(() => {
  document.addEventListener("paste", onPaste);
  document.addEventListener("dragover", onGlobalDragOver);
  document.addEventListener("dragleave", onGlobalDragLeave);
  document.addEventListener("drop", onGlobalDrop);
  document.addEventListener("click", closeEmojiPicker);
});

// Очищаем обработчики при размонтировании
onUnmounted(() => {
  document.removeEventListener("paste", onPaste);
  document.removeEventListener("dragover", onGlobalDragOver);
  document.removeEventListener("dragleave", onGlobalDragLeave);
  document.removeEventListener("drop", onGlobalDrop);
  document.removeEventListener("click", closeEmojiPicker);
});
</script>

<template>
  <!-- Drag & Drop зона -->
  <div
    v-if="isDragOver || isUploading"
    class="drag-zone"
    :class="{ 'drag-over': isDragOver }"
    @dragover="onDragOver"
    @dragleave="onDragLeave"
    @drop="onDrop"
  >
    <div v-if="isDragOver" class="drag-overlay">
      <div class="drag-message">
        📂 Перетащите файлы сюда
        <br />
        <span style="font-size: 16px"
          >📎 Поддерживаются изображения, видео, документы и другие файлы</span
        >
      </div>
    </div>

    <div v-if="isUploading" class="upload-overlay">
      <div class="upload-message">
        ⏳ Загрузка файлов...
        <br />
        <span style="font-size: 16px">📤 Пожалуйста, подождите</span>
      </div>
    </div>
  </div>

  <!-- Боковая панель с комнатами -->
  <div class="sidebar">
    <div class="sidebar-header">
      <h3>🏠 Комнаты</h3>
      <button @click="showCreateRoomModal" class="create-room-btn">➕</button>
    </div>

    <div class="current-room" v-if="currentRoom">
      <div class="current-room-info">
        <div class="room-name">📍 {{ currentRoom.name }}</div>
        <div class="room-users">
          👥 {{ currentRoom.userCount }} пользователей
        </div>
      </div>
    </div>

    <div class="rooms-list">
      <div
        v-for="room in rooms"
        :key="room.id"
        @click="joinRoom(room)"
        :class="['room-item', { active: currentRoom?.id === room.id }]"
      >
        <div class="room-info">
          <div class="room-name">{{ room.name }}</div>
          <div class="room-description">{{ room.description }}</div>
          <div class="room-users">👥 {{ room.userCount }}</div>
        </div>
      </div>
    </div>
  </div>

  <!-- Основная область чата -->
  <div class="chat-area">
    <ul id="messages">
      <li v-for="(msg, i) of messages" :key="i">
        <!-- DEBUG: {{ msg.id }} {{ msg.type }} -->
        <!-- Обычное текстовое сообщение -->
        <template v-if="msg.type === 'text'">
          <div class="message-container">
            <div class="message-content">{{ msg.content }}</div>
            <div class="message-actions">
              <button
                @click.stop="showEmojiPickerForMessage(msg.id)"
                class="emoji-btn-small"
                title="Добавить реакцию"
              >
                😊
              </button>
            </div>
            <!-- Отображение реакций -->
            <div
              v-if="msg.reactions && Object.keys(msg.reactions).length > 0"
              class="reactions"
            >
              <button
                v-for="(reaction, emoji) in msg.reactions"
                :key="emoji"
                @click="addReaction(msg.id, emoji)"
                :class="['reaction-btn', getReactionClass(msg.id, emoji)]"
              >
                {{ emoji }} {{ reaction.count }}
              </button>
            </div>
          </div>
        </template>

        <!-- Превью ссылки -->
        <template v-else-if="msg.type === 'preview'">
          <a
            :href="(msg.content as LinkPreview).url"
            target="_blank"
            class="link-preview"
          >
            <img
              v-if="(msg.content as LinkPreview).image"
              :src="(msg.content as LinkPreview).image"
              :alt="(msg.content as LinkPreview).title"
              class="link-preview-image"
            />
            <div class="link-preview-content">
              <div class="link-preview-title">
                🔗 {{ (msg.content as LinkPreview).title || "Без заголовка" }}
              </div>
              <div
                v-if="(msg.content as LinkPreview).description"
                class="link-preview-description"
              >
                {{ (msg.content as LinkPreview).description }}
              </div>
              <div
                v-if="(msg.content as LinkPreview).siteName"
                class="link-preview-site"
              >
                {{ (msg.content as LinkPreview).siteName }}
              </div>
            </div>
          </a>
        </template>

        <!-- Превью файла -->
        <template v-else-if="msg.type === 'file'">
          <div
            class="file-preview"
            @click="
              openFileModal(
                (msg.content as FilePreview).viewerUrl,
                (msg.content as FilePreview).fileType
              )
            "
          >
            <div class="file-icon">
              <span v-if="(msg.content as FilePreview).fileType === 'pdf'"
                >📄</span
              >
              <span
                v-else-if="(msg.content as FilePreview).fileType === 'document'"
                >📝</span
              >
              <span
                v-else-if="(msg.content as FilePreview).fileType === 'spreadsheet'"
                >📊</span
              >
              <span
                v-else-if="(msg.content as FilePreview).fileType === 'presentation'"
                >📋</span
              >
              <span
                v-else-if="(msg.content as FilePreview).fileType === 'video'"
                >🎬</span
              >
              <span v-else>📎</span>
            </div>
            <div class="file-info">
              <div class="file-name">
                {{ (msg.content as FilePreview).fileName }}
              </div>
              <div class="file-type">
                {{ (msg.content as FilePreview).fileType.toUpperCase() }}
              </div>
            </div>
            <div class="file-action">
              <button class="preview-btn">👁️ Предварительный просмотр</button>
            </div>
          </div>
        </template>

        <!-- Изображение -->
        <template v-else-if="msg.type === 'image'">
          <div
            class="image-message"
            @click="openFileModal(msg.content, 'image')"
          >
            <img :src="msg.content" :alt="msg.content" />
          </div>
        </template>

        <!-- Видео -->
        <template v-else-if="msg.type === 'video'">
          <div
            class="video-message"
            @click="openFileModal(msg.content, 'video')"
          >
            <video controls width="300">
              <source :src="msg.content" />
              Ваш браузер не поддерживает воспроизведение видео.
            </video>
            <div class="video-info">
              🎬 <span>{{ msg.content }}</span>
            </div>
          </div>
        </template>
      </li>
    </ul>
    <form id="form" action="" @submit.prevent="submit">
      <input id="input" autocomplete="off" v-model="message" />
      <input
        type="file"
        ref="fileInput"
        @change="onFileSelect"
        multiple
        style="display: none"
      />
      <button type="button" @click="$refs.fileInput.click()" class="file-btn">
        📎 Файл
      </button>
      <button
        type="button"
        @click.stop="showEmojiPickerForText"
        class="emoji-btn"
      >
        😊
      </button>
      <button type="submit">📤 Отправить</button>
    </form>
  </div>
  <!-- Закрываем chat-area -->

  <!-- Модальное окно для просмотра файлов -->
  <div v-if="isModalOpen" class="modal-overlay" @click="closeModal">
    <div class="modal-content" @click.stop>
      <div class="modal-header">
        <h3>
          <span v-if="modalFileUrl.match(/\.(jpg|jpeg|png|gif|webp|svg)$/i)"
            >🖼️ Просмотр изображения</span
          >
          <span
            v-else-if="
              modalFileUrl.match(/\.(mp4|webm|ogg|avi|mov|wmv|flv|mkv)$/i)
            "
            >🎬 Просмотр видео</span
          >
          <span
            v-else-if="
              modalFileUrl.endsWith('.pdf') || modalFileUrl.includes('pdf')
            "
            >📄 Просмотр PDF</span
          >
          <span
            v-else-if="modalFileUrl.match(/\.(doc|docx|xls|xlsx|ppt|pptx)$/i)"
            >📋 Просмотр документа</span
          >
          <span v-else>📎 Просмотр файла</span>
        </h3>
        <button class="close-btn" @click="closeModal">❌</button>
      </div>
      <div class="modal-body">
        <!-- Показываем PDF через встроенный viewer браузера -->
        <iframe
          v-if="modalFileUrl.endsWith('.pdf') || modalFileUrl.includes('pdf')"
          :src="
            modalFileUrl + '#toolbar=1&navpanes=1&scrollbar=1&page=1&view=FitH'
          "
          width="100%"
          height="100%"
          frameborder="0"
        ></iframe>

        <!-- Показываем изображения -->
        <div
          v-else-if="modalFileUrl.match(/\.(jpg|jpeg|png|gif|webp|svg)$/i)"
          class="image-viewer"
        >
          <img :src="modalFileUrl" alt="Image" class="modal-image" />
        </div>

        <!-- Показываем видео -->
        <div
          v-else-if="
            modalFileUrl.match(/\.(mp4|webm|ogg|avi|mov|wmv|flv|mkv)$/i)
          "
          class="video-viewer"
        >
          <video controls width="100%" height="100%" class="modal-video">
            <source :src="modalFileUrl" />
            Ваш браузер не поддерживает воспроизведение видео.
          </video>
        </div>

        <!-- Для Office документов показываем специальный интерфейс -->
        <div
          v-else-if="
            modalFileUrl.includes('officeapps.live.com') ||
            modalFileUrl.match(/\.(doc|docx|xls|xlsx|ppt|pptx)$/i)
          "
          class="office-viewer"
        >
          <div class="office-info">
            <h4>📋 Просмотр Office документа</h4>
            <p>💡 Если документ не отображается, попробуйте:</p>
            <div class="office-actions">
              <button
                @click="
                  window.open(
                    modalFileUrl.includes('officeapps')
                      ? modalFileUrl.split('src=')[1]
                      : modalFileUrl,
                    '_blank'
                  )
                "
                class="download-btn"
              >
                📥 Скачать файл
              </button>
              <button @click="openInOfficeOnline()" class="online-btn">
                🌐 Открыть в Online Viewer
              </button>
            </div>
          </div>
          <iframe
            :src="modalFileUrl"
            width="100%"
            height="80%"
            frameborder="0"
            @error="onIframeError"
          ></iframe>
        </div>

        <!-- Для неподдерживаемых файлов показываем интерфейс скачивания -->
        <div v-else class="file-download-viewer">
          <div class="download-info">
            <h4>📎 Файл для скачивания</h4>
            <p>
              Этот тип файла не поддерживается для предварительного просмотра.
            </p>
            <div class="download-actions">
              <button
                @click="window.open(modalFileUrl, '_blank')"
                class="download-file-btn"
              >
                📥 Скачать файл
              </button>
              <button
                @click="window.open(modalFileUrl, '_blank')"
                class="open-file-btn"
              >
                🔗 Открыть в новой вкладке
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>

  <!-- Модальное окно создания комнаты -->
  <div v-if="showRoomModal" class="modal-overlay" @click="closeRoomModal">
    <div class="modal-content room-modal" @click.stop>
      <div class="modal-header">
        <h3>🏠 Создать новую комнату</h3>
        <button class="close-btn" @click="closeRoomModal">❌</button>
      </div>
      <div class="modal-body">
        <div class="form-group">
          <label for="roomName">Название комнаты:</label>
          <input
            id="roomName"
            v-model="newRoomName"
            type="text"
            placeholder="Введите название комнаты"
            class="room-input"
            @keyup.enter="createRoom"
          />
        </div>
        <div class="form-group">
          <label for="roomDescription">Описание (необязательно):</label>
          <textarea
            id="roomDescription"
            v-model="newRoomDescription"
            placeholder="Введите описание комнаты"
            class="room-textarea"
            rows="3"
          ></textarea>
        </div>
        <div class="modal-actions">
          <button @click="closeRoomModal" class="cancel-btn">Отмена</button>
          <button @click="createRoom" class="create-btn">
            🏠 Создать комнату
          </button>
        </div>
      </div>
    </div>
  </div>

  <!-- Emoji Picker -->
  <div v-if="showEmojiPicker" class="emoji-picker-popup" @click.stop>
    <div class="quick-reactions">
      <button
        v-for="emoji in ['😊', '😂', '❤️', '👍', '🔥']"
        :key="emoji"
        @click="
          emojiPickerMode === 'text'
            ? addEmojiToText(emoji)
            : addReaction(selectedMessageId, emoji)
        "
        class="quick-emoji"
      >
        {{ emoji }}
      </button>
      <button @click="showMoreEmoji = !showMoreEmoji" class="more-emoji">
        {{ showMoreEmoji ? "▲" : "➕" }}
      </button>
    </div>
    <div v-if="showMoreEmoji" class="more-reactions">
      <button
        v-for="emoji in [
          '😮',
          '😢',
          '😡',
          '🎉',
          '💯',
          '👏',
          '🤔',
          '😍',
          '🙄',
          '😱',
          '💪',
          '🚀',
          '⭐',
          '💎',
          '🎯',
          '👎',
          '😎',
          '🤗',
          '🥳',
        ]"
        :key="emoji"
        @click="
          emojiPickerMode === 'text'
            ? addEmojiToText(emoji)
            : addReaction(selectedMessageId, emoji)
        "
        class="more-emoji-option"
      >
        {{ emoji }}
      </button>
    </div>
  </div>
</template>

<style>
body {
  margin: 0;
  padding: 0;
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica,
    Arial, sans-serif;
  display: flex;
  height: 100vh;
}

/* Боковая панель */
.sidebar {
  width: 300px;
  background: #2c2f33;
  color: white;
  display: flex;
  flex-direction: column;
  border-right: 1px solid #40444b;
}

.sidebar-header {
  padding: 16px;
  background: #23272a;
  display: flex;
  justify-content: space-between;
  align-items: center;
  border-bottom: 1px solid #40444b;
}

.sidebar-header h3 {
  margin: 0;
  font-size: 16px;
}

.create-room-btn {
  background: #7289da;
  border: none;
  color: white;
  width: 32px;
  height: 32px;
  border-radius: 50%;
  cursor: pointer;
  font-size: 18px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background 0.2s;
}

.create-room-btn:hover {
  background: #5b6eae;
}

.current-room {
  padding: 12px 16px;
  background: #36393f;
  border-bottom: 1px solid #40444b;
}

.current-room-info .room-name {
  font-weight: bold;
  margin-bottom: 4px;
}

.current-room-info .room-users {
  font-size: 12px;
  color: #b9bbbe;
}

.rooms-list {
  flex: 1;
  overflow-y: auto;
  padding: 8px 0;
}

.room-item {
  padding: 12px 16px;
  cursor: pointer;
  transition: background 0.2s;
  border-left: 4px solid transparent;
}

.room-item:hover {
  background: #36393f;
}

.room-item.active {
  background: #36393f;
  border-left-color: #7289da;
}

.room-info .room-name {
  font-weight: 500;
  margin-bottom: 4px;
}

.room-info .room-description {
  font-size: 12px;
  color: #72767d;
  margin-bottom: 4px;
}

.room-info .room-users {
  font-size: 11px;
  color: #b9bbbe;
}

/* Основная область чата */
.chat-area {
  flex: 1;
  display: flex;
  flex-direction: column;
  background: white;
}

#form {
  background: rgba(0, 0, 0, 0.15);
  padding: 0.25rem;
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  display: flex;
  height: 3rem;
  box-sizing: border-box;
  backdrop-filter: blur(10px);
}
#input {
  border: none;
  padding: 0 1rem;
  flex-grow: 1;
  border-radius: 2rem;
  margin: 0.25rem;
}
#input:focus {
  outline: none;
}
#form > button {
  background: #333;
  border: none;
  padding: 0 1rem;
  margin: 0.25rem;
  border-radius: 3px;
  outline: none;
  color: #fff;
}

#messages {
  list-style-type: none;
  margin: 0;
  padding: 0;
}
#messages > li {
  padding: 0.5rem 1rem;
}
#messages > li:nth-child(odd) {
  background: #efefef;
}

/* Стили для превью ссылок */
.link-preview {
  border: 1px solid #ddd;
  border-radius: 8px;
  padding: 12px;
  margin: 8px 0;
  display: flex;
  background: #f5f5f5;
  text-decoration: none;
  color: inherit;
  transition: box-shadow 0.2s;
  max-width: 600px;
}

.link-preview:hover {
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.link-preview-image {
  width: 120px;
  height: 90px;
  object-fit: cover;
  border-radius: 4px;
  margin-right: 12px;
  flex-shrink: 0;
}

.link-preview-content {
  flex: 1;
  overflow: hidden;
}

.link-preview-title {
  font-weight: bold;
  margin-bottom: 4px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: #333;
}

.link-preview-description {
  font-size: 14px;
  color: #666;
  overflow: hidden;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  line-height: 1.4;
}

.link-preview-site {
  font-size: 12px;
  color: #999;
  margin-top: 4px;
}

/* Стили для превью файлов */
.file-preview {
  border: 1px solid #ddd;
  border-radius: 8px;
  padding: 16px;
  margin: 8px 0;
  display: flex;
  align-items: center;
  background: #f8f9fa;
  cursor: pointer;
  transition: all 0.2s;
  max-width: 400px;
}

.file-preview:hover {
  background: #e9ecef;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.file-icon {
  font-size: 32px;
  margin-right: 16px;
}

.file-info {
  flex: 1;
}

.file-name {
  font-weight: bold;
  margin-bottom: 4px;
  color: #333;
}

.file-type {
  font-size: 12px;
  color: #666;
  background: #dee2e6;
  padding: 2px 8px;
  border-radius: 12px;
  display: inline-block;
}

.file-action {
  margin-left: 16px;
}

.preview-btn {
  background: #007bff;
  color: white;
  border: none;
  padding: 8px 16px;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
}

.preview-btn:hover {
  background: #0056b3;
}

/* Стили для модального окна */
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background: rgba(0, 0, 0, 0.7);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 1000;
}

.modal-content {
  background: white;
  border-radius: 8px;
  width: 90%;
  max-width: 1000px;
  height: 80%;
  max-height: 700px;
  display: flex;
  flex-direction: column;
}

.modal-header {
  padding: 16px;
  border-bottom: 1px solid #ddd;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.modal-header h3 {
  margin: 0;
  color: #333;
}

.close-btn {
  background: none;
  border: none;
  font-size: 24px;
  cursor: pointer;
  color: #666;
  padding: 0;
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.close-btn:hover {
  color: #333;
  background: #f0f0f0;
  border-radius: 4px;
}

.modal-body {
  flex: 1;
  padding: 0;
  overflow: hidden;
}

.modal-body iframe {
  width: 100%;
  height: 100%;
  border: none;
}

/* Стили для drag & drop */
.drag-zone {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  pointer-events: none;
  z-index: 999;
}

.drag-zone.drag-over {
  pointer-events: all;
  background: rgba(0, 123, 255, 0.1);
}

.drag-overlay {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background: rgba(0, 123, 255, 0.8);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.drag-message {
  background: white;
  padding: 32px;
  border-radius: 16px;
  font-size: 24px;
  font-weight: bold;
  text-align: center;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.3);
}

.upload-overlay {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1001;
}

.upload-message {
  background: white;
  padding: 24px;
  border-radius: 12px;
  font-size: 18px;
  text-align: center;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.3);
}

/* Стили для кнопки файла */
.file-btn {
  background: #28a745;
  color: white;
  border: none;
  padding: 0 12px;
  margin: 0.25rem;
  border-radius: 3px;
  outline: none;
  cursor: pointer;
  font-size: 16px;
}

.file-btn:hover {
  background: #218838;
}

/* Стили для изображений в чате */
.image-message {
  max-width: 300px;
  border-radius: 8px;
  margin: 8px 0;
}

.image-message img {
  width: 100%;
  height: auto;
  border-radius: 8px;
  cursor: pointer;
  transition: transform 0.2s;
}

.image-message img:hover {
  transform: scale(1.02);
}

/* Стили для просмотра изображений в модальном окне */
.image-viewer {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
  background: #f0f0f0;
}

.modal-image {
  max-width: 100%;
  max-height: 100%;
  object-fit: contain;
  border-radius: 4px;
}

/* Улучшенные стили для файлов */
.file-preview .preview-btn {
  transition: all 0.2s;
}

.file-preview .preview-btn:hover {
  transform: translateY(-1px);
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
}

/* Стили для Office viewer */
.office-viewer {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.office-info {
  padding: 16px;
  background: #f8f9fa;
  border-bottom: 1px solid #ddd;
}

.office-info h4 {
  margin: 0 0 8px 0;
  color: #333;
}

.office-info p {
  margin: 0 0 12px 0;
  color: #666;
  font-size: 14px;
}

.office-actions {
  display: flex;
  gap: 12px;
}

.download-btn,
.online-btn {
  padding: 8px 16px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
  transition: all 0.2s;
}

.download-btn {
  background: #28a745;
  color: white;
}

.download-btn:hover {
  background: #218838;
}

.online-btn {
  background: #007bff;
  color: white;
}

.online-btn:hover {
  background: #0056b3;
}

/* Обновляем регулярные выражения для SVG */
.image-message img[src$=".svg"] {
  background: white;
  padding: 8px;
}

/* Стили для видео сообщений */
.video-message {
  max-width: 400px;
  border-radius: 8px;
  margin: 8px 0;
  background: #f8f9fa;
  padding: 8px;
  cursor: pointer;
  transition: transform 0.2s;
}

.video-message:hover {
  transform: scale(1.02);
}

.video-message video {
  width: 100%;
  border-radius: 4px;
}

.video-info {
  padding: 8px 0 4px 0;
  font-size: 14px;
  color: #666;
  text-align: center;
}

/* Стили для видео в модальном окне */
.video-viewer {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
  background: #000;
}

.modal-video {
  max-width: 100%;
  max-height: 100%;
  object-fit: contain;
}

/* Стили для интерфейса скачивания файлов */
.file-download-viewer {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
  background: #f8f9fa;
}

.download-info {
  text-align: center;
  padding: 32px;
  max-width: 400px;
}

.download-info h4 {
  margin: 0 0 16px 0;
  color: #333;
  font-size: 24px;
}

.download-info p {
  margin: 0 0 24px 0;
  color: #666;
  font-size: 16px;
}

.download-actions {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.download-file-btn,
.open-file-btn {
  padding: 12px 24px;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-size: 16px;
  font-weight: 500;
  transition: all 0.2s;
}

.download-file-btn {
  background: #28a745;
  color: white;
}

.download-file-btn:hover {
  background: #218838;
  transform: translateY(-1px);
}

.open-file-btn {
  background: #007bff;
  color: white;
}

.open-file-btn:hover {
  background: #0056b3;
  transform: translateY(-1px);
}

/* Стили для реакций */
.message-container {
  position: relative;
  display: block;
}

.message-content {
  display: inline-block;
  margin-right: 8px;
}

.message-actions {
  display: inline-block;
  opacity: 1;
  transition: opacity 0.2s;
  margin-left: 8px;
  vertical-align: top;
}

.emoji-btn {
  background: #007bff;
  border: 1px solid #007bff;
  cursor: pointer;
  font-size: 18px;
  padding: 6px 10px;
  border-radius: 12px;
  transition: all 0.2s;
  min-width: 40px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
}

.emoji-btn:hover {
  background: #0056b3;
  transform: scale(1.1);
}

.emoji-btn-small {
  background: none;
  border: 1px solid #ddd;
  cursor: pointer;
  font-size: 14px;
  padding: 4px 6px;
  border-radius: 8px;
  transition: all 0.2s;
  min-width: 28px;
  height: 24px;
  display: flex;
  align-items: center;
  justify-content: center;
  opacity: 0.7;
}

.emoji-btn-small:hover {
  background: #f0f0f0;
  opacity: 1;
  transform: scale(1.1);
}

.reactions {
  margin-top: 8px;
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}

.reaction-btn {
  background: #f0f0f0;
  border: 1px solid #ddd;
  border-radius: 12px;
  padding: 4px 8px;
  font-size: 14px;
  cursor: pointer;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  gap: 4px;
}

.reaction-btn:hover {
  background: #e0e0e0;
  transform: scale(1.05);
}

.reaction-active {
  background: #007bff !important;
  color: white;
  border-color: #007bff;
}

.reaction-inactive {
  background: #f0f0f0;
  color: #333;
}

/* Стили для emoji picker */
.emoji-picker-popup {
  position: fixed;
  bottom: 80px;
  left: 50%;
  transform: translateX(-50%);
  background: white;
  border-radius: 20px;
  padding: 8px;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.15);
  z-index: 2000;
  border: 1px solid #e0e0e0;
}

.quick-reactions {
  display: flex;
  gap: 4px;
  align-items: center;
}

.quick-emoji {
  background: none;
  border: none;
  cursor: pointer;
  font-size: 24px;
  padding: 8px 12px;
  border-radius: 16px;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  justify-content: center;
}

.quick-emoji:hover {
  background: #f0f0f0;
  transform: scale(1.1);
}

.more-emoji {
  background: #f0f0f0;
  border: none;
  cursor: pointer;
  font-size: 14px;
  padding: 8px 12px;
  border-radius: 16px;
  transition: all 0.2s;
  color: #666;
  margin-left: 4px;
}

.more-emoji:hover {
  background: #e0e0e0;
}

.more-reactions {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
  margin-top: 8px;
  padding-top: 8px;
  border-top: 1px solid #f0f0f0;
  max-width: 300px;
}

.more-emoji-option {
  background: none;
  border: none;
  cursor: pointer;
  font-size: 20px;
  padding: 6px 10px;
  border-radius: 12px;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  justify-content: center;
}

.more-emoji-option:hover {
  background: #f0f0f0;
  transform: scale(1.1);
}

/* Стили для модального окна создания комнаты */
.room-modal {
  max-width: 500px;
  width: 90%;
}

.form-group {
  margin-bottom: 16px;
}

.form-group label {
  display: block;
  margin-bottom: 8px;
  font-weight: 500;
  color: #333;
}

.room-input,
.room-textarea {
  width: 100%;
  padding: 12px;
  border: 1px solid #ddd;
  border-radius: 6px;
  font-size: 14px;
  box-sizing: border-box;
  transition: border-color 0.2s;
}

.room-input:focus,
.room-textarea:focus {
  outline: none;
  border-color: #7289da;
  box-shadow: 0 0 0 2px rgba(114, 137, 218, 0.2);
}

.room-textarea {
  resize: vertical;
  min-height: 80px;
}

.modal-actions {
  display: flex;
  gap: 12px;
  justify-content: flex-end;
  margin-top: 24px;
}

.cancel-btn,
.create-btn {
  padding: 10px 20px;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-weight: 500;
  transition: all 0.2s;
}

.cancel-btn {
  background: #f0f0f0;
  color: #333;
}

.cancel-btn:hover {
  background: #e0e0e0;
}

.create-btn {
  background: #7289da;
  color: white;
}

.create-btn:hover {
  background: #5b6eae;
  transform: translateY(-1px);
}
</style>
