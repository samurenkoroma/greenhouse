APP=greenhouse

build:
	docker build  -t samurenkoroma/$(APP):0.0.6 .

push:
	docker push samurenkoroma/$(APP):0.0.6