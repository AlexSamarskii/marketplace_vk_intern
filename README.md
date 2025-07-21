# Advertisement Service

Сервис объявлений с регистрацией пользователей, авторизацией и CRUD-операциями для объявлений.  
Проект реализован на Go с использованием `http.ServeMux` для роутинга, PostgreSQL для хранения данных и Docker Compose для локального запуска.

---

## **Структура API**

### **Маршруты `/ad`**
| Метод | Ручка        | Описание |
|-------|--------------|----------|
| `POST` | `/api/v1/ad/create` | Создание нового объявления |
| `GET`  | `/api/v1/ad/{id}`   | Получение информации об объявлении по ID |
| `GET`  | `/api/v1/ad/all`    | Получение списка всех объявлений (с фильтрацией и сортировкой) |

---

### **Маршруты `/auth`**
| Метод | Ручка         | Описание |
|-------|---------------|----------|
| `GET`  | `/api/v1/auth/isAuth`   | Проверка текущей сессии |
| `POST` | `/api/v1/auth/logout`   | Выход из текущей сессии |
| `POST` | `/api/v1/auth/logoutAll`| Выход из всех сессий пользователя |

---

### **Маршруты `/user`**
| Метод | Ручка             | Описание |
|-------|-------------------|----------|
| `POST` | `/api/v1/user/register`   | Регистрация нового пользователя |
| `POST` | `/api/v1/user/login`      | Авторизация (получение токена) |
| `GET`  | `/api/v1/user/profile/{id}` | Получение профиля пользователя по ID |

---

### **Служебные маршруты**
| Метод | Ручка             | Описание |
|-------|-------------------|----------|
| `GET` | `/health`          | Проверка состояния сервиса |
| `GET` | `/swagger/`        | Swagger UI (документация API) |

---
## **Структура .env файла**

Пример `.env` файла:
```env
# PostgreSQL
POSTGRES_HOST=postgres
POSTGRES_CONTAINER_PORT=5432
POSTGRES_HOST_PORT=8070
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
POSTGRES_DB=postgres

# Redis
REDIS_HOST=redis
REDIS_CONTAINER_PORT=6379
REDIS_HOST_PORT=8090
REDIS_PASSWORD=

# App
SERVER_PORT=8000
CSRF_SECRET=9999C55C15065A69AB991BA798A4A498
```

## **Swagger**
Swagger-документация доступна по адресу:
http://localhost:8000/swagger/index.html


Файл спецификации находится в `./docs/swagger.yaml`.

---

## **Запуск проекта**

### **1. Установите Docker и Docker Compose**  
Убедитесь, что у вас установлены:
- [Docker](https://docs.docker.com/get-docker/)
- [Docker Compose](https://docs.docker.com/compose/install/)

---

### **2. Запустите проект**
```bash
docker-compose up --build
```

### **Пример запроса**
```bash
curl -X POST http://localhost:8000/api/v1/ad/create \
-H "Content-Type: application/json" \
-d '{
  "title": "Продам велосипед",
  "description": "Горный велосипед, отличное состояние",
  "price": 10000,
  "image_url": "https://example.com/bike.jpg"
}'
```