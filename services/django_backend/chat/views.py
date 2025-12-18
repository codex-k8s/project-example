from django.http import HttpRequest, HttpResponse


def livez(request: HttpRequest) -> HttpResponse:
    return HttpResponse("ok")


def readyz(request: HttpRequest) -> HttpResponse:
    return HttpResponse("ok")

