from django.contrib import admin
from django.urls import path

from chat import views as chat_views

urlpatterns = [
    path("admin/", admin.site.urls),
    path("health/livez/", chat_views.livez, name="health-livez"),
    path("health/readyz/", chat_views.readyz, name="health-readyz"),
]

