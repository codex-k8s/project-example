from django.contrib import admin

from .models import ChatUser, Message


@admin.register(ChatUser)
class ChatUserAdmin(admin.ModelAdmin):
    list_display = ("id", "nickname", "created_at")
    search_fields = ("nickname",)


@admin.register(Message)
class MessageAdmin(admin.ModelAdmin):
    list_display = ("id", "user", "created_at")
    search_fields = ("user__nickname", "text")
    list_filter = ("created_at",)

