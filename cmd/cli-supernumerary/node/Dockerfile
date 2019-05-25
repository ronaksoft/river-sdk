FROM registry.ronaksoftware.com/base/alpine
WORKDIR /ronak
COPY keys.json /ronak
COPY ./_build/supernumerary /ronak
ENTRYPOINT ["/ronak/supernumerary"]
