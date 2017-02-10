// 原理就是通过 electron (chromium) 提供的 desktop capturer 来获取视频源...
const {desktopCapturer} = require('electron');
const $ = require('jquery');

document.addEventListener('DOMContentLoaded', () => {
  document.querySelector('button[name="start"]')
    .addEventListener('click', start);

  $('#thumbnails').on('click', '.window-thumbnail', function(e) {
    const sourceId = $(this).data('source-id');
    if (!sourceId) {
      console.log(e);
      return;
    }
    selectWindow(sourceId);
  });
});

function start() {
  recordWindow();
}

function recordWindow() {
  desktopCapturer.getSources({ types: ['window'] }, (e, sources) => {
    if (e) throw e;

    // 这里简单显示一下窗口的 thumbnails 吧...
    const thumbnails = document.querySelector('#thumbnails');

    sources.forEach((source) => {
      const thumb = source.thumbnail.toDataURL();
      if (!thumb) return;

      const title = source.name;
      const thumbLi = `<li class="window-thumbnail" data-source-id="${source.id}"><img src="${thumb}" /><span>${title}</span></li>`;
      // 将就将就
      thumbnails.innerHTML += thumbLi;
    });
  });
}

function selectWindow(sourceId) {
  navigator.webkitGetUserMedia({
    // 窗口都没有声音... 得另外截取和混流 o.o
    audio: false,
    video: {
      mandatory: {
        chromeMediaSource: 'desktop',
        chromeMediaSourceId: sourceId,
        maxWidth: 400,
        maxHeight: 300,
      },
    },
  }, getMediaStream, getUseMediaError);
}

function getMediaStream(stream) {
  const video = document.querySelector('#video');
  video.src = URL.createObjectURL(stream);

  const recorder = new MediaRecorder(stream);
  recorder.ondataavailable = (d) => {
    // 在这里截获视频流
    //
    // TODO 接下来会在这里把视频流发送到服务器端
  };
  recorder.start();
}

function getUseMediaError(e) {
  console.log(`${e} failed`);
}
