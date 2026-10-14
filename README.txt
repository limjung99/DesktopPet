데스크톱 펫 (Windows 10/11, 설치 불필요)

[실행]
  DesktopPet.exe 더블클릭. Python 등 다른 프로그램 필요 없음.
  중복 실행은 자동으로 막힙니다.

[조작]
  왼쪽 클릭   : 점프 + 말풍선
  드래그      : 집어서 던지기
  오른쪽 클릭 : 메뉴 (말하기 / 점프 / 종료)

[다중 모니터]
  연결된 모든 모니터를 오가며 걸어다닙니다.
  - 바닥이 더 낮은 모니터로 넘어가면 떨어지고, 더 높은 쪽(작업표시줄이 있는 모니터 등)은 점프해서 올라갑니다.
  - 모니터 사이가 떨어져 있으면 끝에서 돌아서고, 드래그해서 옮길 수는 있습니다.
  - 모니터 연결/해제, 해상도·작업표시줄 변경은 자동 반영됩니다.

[설정] (exe와 같은 폴더)
  pet.ini    : 크기(scale), 속도, 혼잣말 확률
  lines.txt  : 말풍선 대사 (한 줄에 하나)
  sprites\   : 이 폴더를 만들고 idle_0.png, walk_0.png ... 를 넣으면
               내장 캐릭터 대신 그 이미지를 사용 (src\sprites 참고)

[시작 시 자동 실행]
  Win+R -> shell:startup -> 열린 폴더에 DesktopPet.exe 바로가기 넣기

[다시 빌드하려면] (src 폴더, Go 1.21 이상)
  set GOOS=windows& go build -ldflags "-H windowsgui -s -w" -o DesktopPet.exe .
