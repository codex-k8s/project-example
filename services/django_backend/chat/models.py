from django.db import models


class ChatUser(models.Model):
    nickname = models.CharField(max_length=64, unique=True, db_index=True)
    password_hash = models.CharField(max_length=128)
    created_at = models.DateTimeField(auto_now_add=True)

    class Meta:
        db_table = "chat_user"
        ordering = ["id"]

    def __str__(self) -> str:
        return self.nickname


class Message(models.Model):
    user = models.ForeignKey(ChatUser, on_delete=models.CASCADE, related_name="messages")
    text = models.TextField()
    created_at = models.DateTimeField(auto_now_add=True, db_index=True)

    class Meta:
        db_table = "chat_message"
        ordering = ["id"]

    def __str__(self) -> str:
        return f"{self.user.nickname}: {self.text[:40]}"

