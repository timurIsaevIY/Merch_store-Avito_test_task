# 🛍 Merch Store - Avito Test Task

## 🚀 Запуск проекта
Для развертывания и запуска проекта используйте:
``docker-compose up -d --build``
# ✅ Тестирование

## 🚀 Запуск тестов

<table>
  <thead>
    <tr>
      <th align="left">Тип тестов</th>
      <th align="left">Команда запуска</th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <td>🚧 <strong>Интеграционные</strong></td>
      <td><code>make integration-test</code></td>
    </tr>
    <tr>
      <td>⚙️ <strong>Unit-тесты с покрытием</strong></td>
      <td><code>make unit-test-cover</code></td>
    </tr>
    <tr>
      <td>🔥 <strong>Нагрузочное тестирование</strong></td>
      <td><code>make load-test</code></td>
    </tr>
  </tbody>
</table>

## 📊 Тестовое покрытие

Текущее покрытие кода тестами: **49.3%**

![Тестовое покрытие](https://github.com/user-attachments/assets/61b395e7-583a-4afe-a031-7174fdec9d03)

## 🔥 Нагрузочное тестирование

Результаты нагрузочного тестирования:

![Нагрузочное тестирование](https://github.com/user-attachments/assets/f04d5379-2f22-4bad-b50b-98d356e40f7f)

---

### ℹ️ Примечания:
1. Файл `.env` оставлен намеренно для быстрого тестирования проекта.
2. Проект разрабатывался из-под **Windows**, поэтому некоторые команды в `Makefile` могут не работать на **Linux/MacOS**.
