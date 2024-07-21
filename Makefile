APP=greenhouse

build:
	docker build  -t samurenkoroma/$(APP):0.0.13 .

push:
	docker push samurenkoroma/$(APP):0.0.13