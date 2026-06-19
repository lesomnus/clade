관심있는 컨테이너 이미지들 태그를 조회해서 그 태그의 이미지를 베이스로 하는 도커파일을 빌드한 뒤 내가 원하는 레지스트리에 푸쉬하는 과정을 자동화하고싶어.
CI에서 cron으로 업스트림 이미지를 검사하고 내가 정의한 목적지 이미지랑 비교해서 내 이미지가 오래됐으면 다시 빌드가 돌아가서 내 이미지들의 베이스를 항상 최신으로 유지하는거지.
그럼 outdated 명령어로 내 이미지중 upstream과 최신이 아닌 이미지 목록을 볼 수 있어야겠지.
참고로 내 이미지 자체도 다시 업스트림이 되서 하위의 다른 내 이미지가 다시 빌드 대상이 될거야.
그래서 정확히는 outdated 빌드 대상이 되는 이미지들의 그래프를 출력하겠지.
그래프는 proto로 정의해서 serialize될 수 있으면 좋겠어.
빌드 그래프를 가지고 build명령을 실행하면 해당되는 도커파일에 BASE 인자로 업스트림 이미지 이름과 함께 차례대로 빌드가 되는거야.

일단 내가 빌드하고자 하는 Dockerfile은 context와 함께 디렉터리 안에 있을거야.
내가 빌드하고자 하는 이미지들을 port라고하자. 기본적으로 port들은 ports라는 디렉터리에 있어:
```
ports/
  golang-dev/
	Dockerfile
	somefile
	port.yaml
  golang-dev-alpine/
	Dockerfile
	somefile
	port.yaml
  ...
```

port.yaml 에는 업스트림 이미지 이름과 태그 패턴이 정의되어 있어야 해.
```
parent:
  repo: docker.io/library/golang
  target:
    kind: semver
	last-major: 2  # Select latest two major versions
	last-minor: 3  # Select latest three minor versions
	match: "_-alpine"  # Select tags that end with "-alpine"

build:
  repo: my-registry/my-image/golang-dev
  tag: "{{.Major}}.{{.Minor}}.{{.Patch}}-alpine"
```
이런 느낌으로 내 이미지 빌드를 정의할 수 있지 않을까?
tag가 semver만 있는건 아니니까 kind는 추가로 정의될 수 있도록 target tag를 검색하는건 인터페이스로 추상화 되어 있어야할 것 같아.

예를 들어 업스트림에 `docker.io/library/golang:1.2.3-alpine` 이미지가 있으면 내 레지스트리에 `my-registry/my-image/golang-dev:1.2.3-alpine` 이미지가 있는지 확인해서 없거나 오래된 경우에 빌드가 돌아가도록 하는거지.
오래된 것을 판단하는 것도 인터페이스로 추상화됐으면 좋겠어. 예를 들어 빌드된 이미지의 생성 날짜와 업스트림 이미지의 생성 날짜를 비교해서 오래된 것을 판단할 수도 있고, 빌드된 이미지의 digest와 업스트림 이미지의 digest를 비교해서 오래된 것을 판단할 수도 있겠지.
향후 누군가는 라벨을 비교한다거나 하는 요구가 있을 수 있으니까 그것을 고려해서 오래된 것을 판단하는 로직도 인터페이스로 추상화되어야 할 것 같아.

이 앱을 일반적으로 생각한다면 업스트림 이미지와 타겟 이미지 사이의 관계를 나타내서 그래프로 표현하고 업스트림이 업데이트되면 업스트림 이미지와 타겟 이미지 정보를 인자로 하는 트리거 시스템인거야.
이 앱은 트리거되면 빌드를 수행하는 거고.

그래서 일단 그래프를 구성하고 outdated된 이미지를 확인하는 로직부터 구현해야해.
docker는 메타데이터 확인만으로도 quota가 소진되니까 업스트림 이미지와 타겟 이미지의 메타데이터를 캐싱하는 시스템이 필요할 것 같고, 로컬 테스트를 위해선 registry mock이 필요할 것 같아.
distribution/distribution 라이브러리를 어차피 사용하겠지? 거기 있는 클라이언트 인터페이스를 mocking하면 되려나?

구현 계획을 세워줘. port 규격이나 그래프 proto 규격은 필요에따라 내 정의를 수정해도 되. 구현 계획을 세워줘.
