<script setup>
import { computed, nextTick, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue';

const API_BASE = (import.meta.env.VITE_API_BASE_URL || '').replace(/\/$/, '');
const STORAGE_TOKEN = 'my-feed-token';
const STORAGE_ACCOUNT = 'my-feed-account';

const token = ref(localStorage.getItem(STORAGE_TOKEN) || '');
const account = ref(readStoredAccount());
const appPage = ref('feed');
const currentView = ref('latest');
const activeVideoId = ref(null);
const selectedVideo = ref(null);
const detailLoading = ref(false);
const videoPanelOpen = ref(false);
const videoPanelTab = ref('author');
const relationDrawer = ref(false);
const reelScroller = ref(null);
const toast = reactive({ type: '', text: '' });
const videoRefs = new Map();
let reelObserver = null;

const VIDEO_PANEL_TABS = [
  { id: 'author', label: '作者' },
  { id: 'comments', label: '评论' },
  { id: 'messages', label: '私信' }
];

const player = reactive({
  duration: 0,
  currentTime: 0,
  paused: true,
  muted: false
});

const authMode = ref('login');
const authForm = reactive({ username: '', password: '' });
const authLoading = ref(false);
const meLoading = ref(false);

const latestFeed = reactive({
  videos: [],
  latest_time: 0,
  latest_id: 0,
  next_time: 0,
  next_id: 0,
  has_more: false,
  loading: false,
  error: ''
});

const tagFeed = reactive({
  tag_name: '',
  activeTag: '',
  videos: [],
  latest_time: 0,
  latest_id: 0,
  next_time: 0,
  next_id: 0,
  has_more: false,
  loading: false,
  error: ''
});

const authorFeed = reactive({
  author_id: null,
  videos: [],
  latest_time: 0,
  latest_id: 0,
  next_time: 0,
  next_id: 0,
  has_more: false,
  loading: false,
  error: ''
});

const likedFeed = reactive({
  videos: [],
  latest_time: 0,
  latest_id: 0,
  next_time: 0,
  next_id: 0,
  has_more: false,
  loading: false,
  error: ''
});

const publishForm = reactive({
  title: '',
  description: '',
  play_url: '',
  cover_url: ''
});

const uploadState = reactive({
  videoFile: null,
  coverFile: null,
  videoPreview: '',
  coverPreview: '',
  videoUploading: false,
  coverUploading: false,
  publishing: false,
  error: ''
});

const commentState = reactive({
  videoId: null,
  videoTitle: '',
  comments: [],
  draft: '',
  latest_time: 0,
  latest_id: 0,
  next_time: 0,
  next_id: 0,
  has_more: false,
  loading: false,
  publishing: false,
  deletingById: {},
  error: ''
});

const relationState = reactive({
  mode: 'followers',
  accountId: null,
  title: '',
  accounts: [],
  latest_time: 0,
  latest_id: 0,
  next_time: 0,
  next_id: 0,
  has_more: false,
  loading: false,
  error: ''
});

const messageState = reactive({
  toId: null,
  title: '',
  messages: [],
  draft: '',
  latest_time: 0,
  latest_id: 0,
  next_time: 0,
  next_id: 0,
  has_more: false,
  loading: false,
  sending: false,
  error: ''
});

const likedById = reactive({});
const likeLoadingById = reactive({});
const followingById = reactive({});
const followLoadingById = reactive({});
const profileById = reactive({});
const profileLoadingById = reactive({});

const isAuthed = computed(() => Boolean(token.value));
const visibleVideos = computed(() => {
  if (currentView.value === 'tag') return tagFeed.videos;
  if (currentView.value === 'liked') return likedFeed.videos;
  if (currentView.value === 'author') return authorFeed.videos;
  return latestFeed.videos;
});
const activeFeed = computed(() => {
  if (currentView.value === 'tag') return tagFeed;
  if (currentView.value === 'liked') return likedFeed;
  if (currentView.value === 'author') return authorFeed;
  return latestFeed;
});
const selectedAccountName = computed(() => account.value?.username || '未登录');
const feedTitle = computed(() => {
  if (currentView.value === 'tag' && tagFeed.activeTag) return `#${tagFeed.activeTag}`;
  if (currentView.value === 'liked') return '已赞视频';
  if (currentView.value === 'author') return `${profileName(authorFeed.author_id, `作者 ${authorFeed.author_id}`)}的作品`;
  return '最新视频';
});
const myProfile = computed(() => (account.value?.id ? profileById[account.value.id] : null));
const authorProfile = computed(() => (authorFeed.author_id ? profileById[authorFeed.author_id] : null));
const videoPanelHeading = computed(() => {
  if (videoPanelTab.value === 'comments') return '评论区';
  if (videoPanelTab.value === 'messages') return '私信';
  const authorId = authorFeed.author_id || activeVideo.value?.author_id;
  return authorId ? profileName(authorId, `作者 ${authorId}`) : '作者';
});
const videoPanelSubtitle = computed(() => {
  if (videoPanelTab.value === 'comments') return commentState.videoTitle || activeVideo.value?.title || '当前视频';
  if (videoPanelTab.value === 'messages') return messageState.title || '当前作者';
  const authorId = authorFeed.author_id || activeVideo.value?.author_id;
  return authorId ? `ID ${authorId}` : '当前视频';
});
const hasUploadedVideo = computed(() => Boolean(publishForm.play_url.trim()));
const hasUploadedCover = computed(() => Boolean(publishForm.cover_url.trim()));
const publishReady = computed(() => isAuthed.value && publishForm.title.trim() && hasUploadedVideo.value);
const progressPercent = computed(() => {
  if (!player.duration) return '0%';
  const percent = Math.min(100, Math.max(0, (player.currentTime / player.duration) * 100));
  return `${percent}%`;
});

const activeVideo = computed(() => {
  if (!visibleVideos.value.length) return null;
  return visibleVideos.value.find((video) => videoId(video) === activeVideoId.value) || visibleVideos.value[0];
});

onMounted(async () => {
  await loadLatest(true);
  if (token.value) {
    await refreshMe(false);
  }
  await nextTick();
  setupReelObserver();
});

watch(activeVideoId, async () => {
  const shouldResumePlayback = !player.paused;
  resetPlayer();
  await nextTick();
  pauseInactiveVideos();
  const mediaElement = activeVideoElement();
  syncPlayerState(mediaElement);
  if (shouldResumePlayback && mediaElement) {
    try {
      await mediaElement.play();
      syncPlayerState(mediaElement);
    } catch {
      syncPlayerState(mediaElement);
    }
  }
  await loadActiveLikeState();
  await loadActiveAuthorFollowState();
  await loadActiveAuthorProfile();
  await syncVideoPanelWithActive();
});

watch(
  () => visibleVideos.value.map((video) => videoId(video)).join(','),
  async () => {
    await nextTick();
    setupReelObserver();
    const activeStillVisible = visibleVideos.value.some((video) => videoId(video) === activeVideoId.value);
    if (visibleVideos.value.length && (!activeVideoId.value || !activeStillVisible)) {
      setActiveVideo(visibleVideos.value[0]);
    }
  },
  { flush: 'post' }
);

watch(appPage, async (page) => {
  if (page === 'feed') {
    await nextTick();
    setupReelObserver();
  } else {
    disconnectReelObserver();
  }
});

onBeforeUnmount(() => {
  disconnectReelObserver();
});

function readStoredAccount() {
  try {
    return JSON.parse(localStorage.getItem(STORAGE_ACCOUNT) || 'null');
  } catch {
    return null;
  }
}

function showToast(text, type = 'success') {
  toast.text = text;
  toast.type = type;
  window.clearTimeout(showToast.timer);
  showToast.timer = window.setTimeout(() => {
    toast.text = '';
    toast.type = '';
  }, 2800);
}

function goPage(page) {
  appPage.value = page;
  if (page === 'feed' && !visibleVideos.value.length) {
    loadLatest(true);
  }
}

async function apiRequest(path, options = {}) {
  const { auth = false, form = false, body, ...fetchOptions } = options;
  const headers = new Headers(fetchOptions.headers || {});

  if (!form) {
    headers.set('Content-Type', 'application/json');
  }
  if (auth) {
    if (!token.value) {
      throw new Error('请先登录');
    }
    headers.set('Authorization', `Bearer ${token.value}`);
  }

  const requestInit = {
    ...fetchOptions,
    headers
  };

  if (body !== undefined) {
    requestInit.body = form || typeof body === 'string' ? body : JSON.stringify(body);
  }

  const response = await fetch(`${API_BASE}${path}`, requestInit);
  const contentType = response.headers.get('content-type') || '';
  const data = contentType.includes('application/json') ? await response.json() : await response.text();

  if (!response.ok) {
    if (response.status === 401) {
      clearSession();
    }
    const message = typeof data === 'string' ? data : data?.error || data?.message || '请求失败';
    throw new Error(message);
  }

  return data;
}

function normalizeUrl(url) {
  if (!url) return '';
  if (/^https?:\/\//i.test(url)) return url;
  return `${API_BASE}${url.startsWith('/') ? '' : '/'}${url}`;
}

function videoId(video) {
  return recordId(video);
}

function recordId(record) {
  return record?.id ?? record?.ID;
}

function createdAtOf(record) {
  return record?.created_at ?? record?.CreatedAt;
}

function formatTime(value) {
  if (!value) return '刚刚';
  const numericValue = typeof value === 'number' ? value : /^\d+$/.test(String(value)) ? Number(value) : null;
  const normalized = numericValue === null ? String(value).replace(/\.(\d{3})\d+/, '.$1') : numericValue;
  const date = new Date(normalized);
  if (Number.isNaN(date.getTime())) return '刚刚';
  return new Intl.DateTimeFormat('zh-CN', {
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  }).format(date);
}

function formatDuration(value) {
  const totalSeconds = Math.max(0, Math.floor(Number(value) || 0));
  const minutes = Math.floor(totalSeconds / 60);
  const seconds = String(totalSeconds % 60).padStart(2, '0');
  return `${minutes}:${seconds}`;
}

function compactNumber(value) {
  const number = Number(value || 0);
  if (number >= 10000) return `${(number / 10000).toFixed(1)}w`;
  return String(number);
}

function extractTags(video) {
  const source = `${video?.title || ''} ${video?.description || ''}`;
  const matches = source.match(/#[\p{L}\p{N}_-]+/gu) || [];
  return [...new Set(matches.map((tag) => tag.toLowerCase()))];
}

function setActiveVideo(video) {
  if (!video) return;
  activeVideoId.value = videoId(video);
}

function playPanelVideo(video) {
  if (!video) return;
  currentView.value = 'author';
  appPage.value = 'feed';
  setActiveVideo(video);
}

function closeVideoPanel() {
  videoPanelOpen.value = false;
}

async function switchVideoPanelTab(tab) {
  videoPanelOpen.value = true;
  videoPanelTab.value = tab;
  await syncVideoPanelWithActive({ notify: true });
}

async function openAuthorMessages(authorId = authorFeed.author_id) {
  if (!authorId) return;
  await openMessages(authorId, profileName(authorId, `作者 ${authorId}`));
}

async function openAuthorPanel(authorId = activeVideo.value?.author_id) {
  if (!authorId) return;
  appPage.value = 'feed';
  videoPanelOpen.value = true;
  videoPanelTab.value = 'author';
  await loadAuthorVideos(authorId, true);
}

async function syncVideoPanelWithActive(options = {}) {
  if (!videoPanelOpen.value) return;
  const video = activeVideo.value;
  if (!video) return;

  if (videoPanelTab.value === 'author') {
    if (authorFeed.author_id !== video.author_id || !authorFeed.videos.length) {
      await loadAuthorVideos(video.author_id, true);
    }
    return;
  }

  if (videoPanelTab.value === 'comments') {
    await prepareComments(video);
    return;
  }

  if (videoPanelTab.value === 'messages') {
    await prepareMessagesForVideo(video, options);
  }
}

function resetCursor(target) {
  target.latest_time = 0;
  target.latest_id = 0;
  target.next_time = 0;
  target.next_id = 0;
  target.has_more = false;
}

function cursorBody(target) {
  return {
    latest_time: target.latest_time || 0,
    latest_id: target.latest_id || 0
  };
}

function applyCursor(target, data) {
  target.next_time = data?.next_time || 0;
  target.next_id = data?.next_id || 0;
  target.latest_time = target.next_time;
  target.latest_id = target.next_id;
  target.has_more = Boolean(data?.has_more);
}

function isActiveVideo(video) {
  return videoId(video) === activeVideoId.value;
}

function videoRefKey(videoOrId) {
  const id = typeof videoOrId === 'object' ? videoId(videoOrId) : videoOrId;
  return id === undefined || id === null ? '' : String(id);
}

function setVideoRef(video, element) {
  const key = videoRefKey(video);
  if (!key) return;
  if (element) {
    videoRefs.set(key, element);
  } else {
    videoRefs.delete(key);
  }
}

function activeVideoElement() {
  return videoRefs.get(videoRefKey(activeVideoId.value)) || null;
}

function pauseInactiveVideos() {
  const activeKey = videoRefKey(activeVideoId.value);
  for (const [key, element] of videoRefs.entries()) {
    if (key !== activeKey && !element.paused) {
      element.pause();
    }
  }
}

function disconnectReelObserver() {
  if (reelObserver) {
    reelObserver.disconnect();
    reelObserver = null;
  }
}

function setupReelObserver() {
  disconnectReelObserver();
  const scroller = reelScroller.value;
  if (!scroller) return;
  const items = Array.from(scroller.querySelectorAll('[data-reel-video-id]'));
  if (!items.length) return;

  reelObserver = new IntersectionObserver(
    (entries) => {
      const visibleEntry = entries
        .filter((entry) => entry.isIntersecting)
        .sort((left, right) => right.intersectionRatio - left.intersectionRatio)[0];
      if (!visibleEntry || visibleEntry.intersectionRatio < 0.55) return;

      const id = visibleEntry.target.getAttribute('data-reel-video-id');
      const video = visibleVideos.value.find((item) => videoRefKey(item) === id);
      if (video && !isActiveVideo(video)) {
        setActiveVideo(video);
      }

      const index = visibleVideos.value.findIndex((item) => videoRefKey(item) === id);
      if (index >= visibleVideos.value.length - 2 && activeFeed.value.has_more && !activeFeed.value.loading) {
        loadMoreActiveFeed();
      }
    },
    {
      root: scroller,
      threshold: [0.55, 0.7, 0.85]
    }
  );

  for (const item of items) {
    reelObserver.observe(item);
  }
}

function loadMoreActiveFeed() {
  if (currentView.value === 'tag') return loadByTag(false);
  if (currentView.value === 'liked') return loadLikedVideos(false);
  if (currentView.value === 'author') return loadAuthorVideos(authorFeed.author_id, false);
  return loadLatest(false);
}

function profileAccount(id) {
  return profileById[id]?.account || null;
}

function profileName(id, fallback = '') {
  return profileAccount(id)?.username || fallback || `用户 ${id}`;
}

function profileAvatar(id, fallback = '') {
  return profileAccount(id)?.avatar_url || fallback || '';
}

function profileBio(id, fallback = '') {
  return profileAccount(id)?.bio || fallback || '';
}

function cacheProfile(data) {
  const profile = data?.account ? data : null;
  const id = profile?.account?.id;
  if (!id) return null;
  profileById[id] = {
    account: profile.account,
    stats: profile.stats || {}
  };
  if (account.value?.id === id) {
    account.value = { ...account.value, ...profile.account };
    localStorage.setItem(STORAGE_ACCOUNT, JSON.stringify(account.value));
  }
  return profileById[id];
}

async function loadProfile(accountId, force = false) {
  if (!accountId) return null;
  if (!force && profileById[accountId]) return profileById[accountId];
  if (profileLoadingById[accountId]) return profileById[accountId] || null;

  profileLoadingById[accountId] = true;
  try {
    const data = await apiRequest('/account/getProfile', {
      method: 'POST',
      body: { account_id: accountId }
    });
    return cacheProfile(data);
  } catch (error) {
    return null;
  } finally {
    profileLoadingById[accountId] = false;
  }
}

function isVideoLiked(video) {
  return Boolean(likedById[videoId(video)]);
}

function isLikeLoading(video) {
  return Boolean(likeLoadingById[videoId(video)]);
}

function setVideoLiked(id, value) {
  if (!id) return;
  likedById[id] = Boolean(value);
}

function isOwnAccount(id) {
  return Boolean(isAuthed.value && account.value?.id === id);
}

function isFollowing(id) {
  return Boolean(followingById[id]);
}

function isFollowLoading(id) {
  return Boolean(followLoadingById[id]);
}

function setFollowing(id, value) {
  if (!id) return;
  followingById[id] = Boolean(value);
}

function updateVideoLikes(id, likesCount) {
  if (!id || likesCount === undefined || likesCount === null) return;
  const feeds = [latestFeed, tagFeed, likedFeed, authorFeed];
  for (const feed of feeds) {
    const video = feed.videos.find((item) => videoId(item) === id);
    if (video) {
      video.likes_count = likesCount;
    }
  }
  if (selectedVideo.value && videoId(selectedVideo.value) === id) {
    selectedVideo.value.likes_count = likesCount;
  }
}

function updateVideoComments(id, commentsCount) {
  if (!id || commentsCount === undefined || commentsCount === null) return;
  const feeds = [latestFeed, tagFeed, likedFeed, authorFeed];
  for (const feed of feeds) {
    const video = feed.videos.find((item) => videoId(item) === id);
    if (video) {
      video.comments_count = commentsCount;
    }
  }
  if (selectedVideo.value && videoId(selectedVideo.value) === id) {
    selectedVideo.value.comments_count = commentsCount;
  }
}

async function loadActiveLikeState() {
  const id = activeVideoId.value;
  if (!id || !isAuthed.value) return;
  if (likedById[id] !== undefined) return;
  try {
    const data = await apiRequest('/like/isLiked', {
      method: 'POST',
      auth: true,
      body: { video_id: id }
    });
    setVideoLiked(id, data?.is_liked);
  } catch (error) {
    if (token.value) showToast(error.message, 'error');
  }
}

async function loadActiveAuthorFollowState() {
  const authorId = activeVideo.value?.author_id;
  if (!authorId || !isAuthed.value || isOwnAccount(authorId)) return;
  if (followingById[authorId] !== undefined) return;
  await loadFollowState(authorId);
}

async function loadActiveAuthorProfile() {
  const authorId = activeVideo.value?.author_id;
  if (!authorId) return;
  await loadProfile(authorId);
}

async function loadFollowState(vloggerId) {
  if (!vloggerId || !isAuthed.value || isOwnAccount(vloggerId)) return;
  try {
    const data = await apiRequest('/follow/isFollowing', {
      method: 'POST',
      auth: true,
      body: { vlogger_id: vloggerId }
    });
    setFollowing(vloggerId, data?.is_following);
  } catch (error) {
    if (token.value) showToast(error.message, 'error');
  }
}

async function toggleFollow(vloggerId) {
  if (!vloggerId) return;
  if (!isAuthed.value) {
    showToast('请先登录后关注', 'error');
    appPage.value = 'account';
    return;
  }
  if (isOwnAccount(vloggerId)) {
    showToast('不能关注自己', 'error');
    return;
  }
  if (followLoadingById[vloggerId]) return;

  const currentlyFollowing = isFollowing(vloggerId);
  followLoadingById[vloggerId] = true;
  try {
    await apiRequest(currentlyFollowing ? '/follow/unfollow' : '/follow/follow', {
      method: 'POST',
      auth: true,
      body: { vlogger_id: vloggerId }
    });
    setFollowing(vloggerId, !currentlyFollowing);
    showToast(currentlyFollowing ? '已取消关注' : '已关注');
    if (currentlyFollowing && relationState.mode === 'following' && relationState.accountId === account.value?.id) {
      relationState.accounts = relationState.accounts.filter((item) => recordId(item) !== vloggerId);
    }
  } catch (error) {
    showToast(error.message, 'error');
  } finally {
    followLoadingById[vloggerId] = false;
  }
}

async function openRelations(mode, accountId, title = '') {
  if (!accountId) return;
  relationDrawer.value = true;
  relationState.title = title || `用户 ${accountId}`;
  if (relationState.mode !== mode || relationState.accountId !== accountId) {
    relationState.mode = mode;
    relationState.accountId = accountId;
    relationState.accounts = [];
    resetCursor(relationState);
    relationState.error = '';
  }
  const profile = await loadProfile(accountId);
  if (profile?.account?.username) {
    relationState.title = profile.account.username;
  }
  await loadRelations(true);
}

async function loadRelations(reset = false) {
  if (!relationState.accountId || relationState.loading) return;
  relationState.loading = true;
  relationState.error = '';
  try {
    if (reset) {
      resetCursor(relationState);
      relationState.accounts = [];
    }
    const data = await apiRequest(relationState.mode === 'followers' ? '/follow/listFollower' : '/follow/listFollowing', {
      method: 'POST',
      body: {
        account_id: relationState.accountId,
        limit: 20,
        ...cursorBody(relationState)
      }
    });
    const accounts = Array.isArray(data?.accounts) ? data.accounts : [];
    relationState.accounts = reset ? accounts : [...relationState.accounts, ...accounts];
    applyCursor(relationState, data);
    if (relationState.mode === 'following' && relationState.accountId === account.value?.id) {
      for (const item of relationState.accounts) {
        setFollowing(recordId(item), true);
      }
    }
  } catch (error) {
    relationState.error = error.message;
  } finally {
    relationState.loading = false;
  }
}

function messageTimeValue(message) {
  const time = new Date(createdAtOf(message) || '').getTime();
  return Number.isNaN(time) ? 0 : time;
}

function normalizeConversationMessages(messages) {
  return [...messages].sort((left, right) => messageTimeValue(left) - messageTimeValue(right));
}

function isOwnMessage(message) {
  return Boolean(account.value?.id && message?.from_id === account.value.id);
}

async function openMessages(toId, title = '') {
  if (!toId) return;
  if (!isAuthed.value) {
    showToast('请先登录后私信', 'error');
    appPage.value = 'account';
    return;
  }
  if (isOwnAccount(toId)) {
    showToast('不能给自己发私信', 'error');
    return;
  }

  appPage.value = 'feed';
  relationDrawer.value = false;
  videoPanelOpen.value = true;
  videoPanelTab.value = 'messages';
  await prepareMessageConversation(toId, title);
}

async function prepareMessagesForVideo(video, options = {}) {
  const toId = video?.author_id;
  if (!toId) return;
  if (!isAuthed.value) {
    if (options.notify) {
      showToast('请先登录后私信', 'error');
      appPage.value = 'account';
    }
    return;
  }
  if (isOwnAccount(toId)) {
    messageState.toId = null;
    messageState.title = '自己的视频';
    messageState.messages = [];
    messageState.draft = '';
    resetCursor(messageState);
    messageState.error = '不能给自己发私信';
    if (options.notify) showToast('不能给自己发私信', 'error');
    return;
  }

  await prepareMessageConversation(toId, profileName(toId, `作者 ${toId}`));
}

async function prepareMessageConversation(toId, title = '') {
  if (!toId) return;
  messageState.toId = toId;
  messageState.title = title || `用户 ${toId}`;
  messageState.messages = [];
  messageState.draft = '';
  resetCursor(messageState);
  messageState.error = '';
  const profile = await loadProfile(toId);
  if (profile?.account?.username) {
    messageState.title = profile.account.username;
  }
  await loadConversation(true);
}

async function loadConversation(reset = false) {
  if (!messageState.toId || messageState.loading) return;
  messageState.loading = true;
  messageState.error = '';
  try {
    if (reset) {
      resetCursor(messageState);
      messageState.messages = [];
    }
    const data = await apiRequest('/message/listConversation', {
      method: 'POST',
      auth: true,
      body: {
        to_id: messageState.toId,
        limit: 20,
        ...cursorBody(messageState)
      }
    });
    const messages = normalizeConversationMessages(Array.isArray(data?.messages) ? data.messages : []);
    messageState.messages = reset ? messages : [...messages, ...messageState.messages];
    applyCursor(messageState, data);
  } catch (error) {
    messageState.error = error.message;
  } finally {
    messageState.loading = false;
  }
}

async function sendMessage() {
  const content = messageState.draft.trim();
  if (!messageState.toId || messageState.sending) return;
  if (!content) {
    messageState.error = '请输入私信内容';
    return;
  }

  messageState.sending = true;
  messageState.error = '';
  try {
    const data = await apiRequest('/message/sendMsg', {
      method: 'POST',
      auth: true,
      body: {
        to_id: messageState.toId,
        content
      }
    });
    if (data?.message) {
      messageState.messages = [...messageState.messages, data.message];
    }
    messageState.draft = '';
    showToast('私信已发送');
  } catch (error) {
    messageState.error = error.message;
    showToast(error.message, 'error');
  } finally {
    messageState.sending = false;
  }
}

async function openComments(video) {
  const id = videoId(video);
  if (!id) return;
  appPage.value = 'feed';
  videoPanelOpen.value = true;
  videoPanelTab.value = 'comments';
  await prepareComments(video);
}

async function prepareComments(video) {
  const id = videoId(video);
  if (!id) return;
  if (commentState.videoId !== id) {
    commentState.videoId = id;
    commentState.videoTitle = video.title || `视频 ${id}`;
    commentState.comments = [];
    commentState.draft = '';
    resetCursor(commentState);
    commentState.error = '';
    await loadComments(true);
  }
}

async function loadComments(reset = false) {
  if (!commentState.videoId || commentState.loading) return;
  commentState.loading = true;
  commentState.error = '';
  try {
    if (reset) {
      resetCursor(commentState);
      commentState.comments = [];
    }
    const data = await apiRequest('/comment/listComment', {
      method: 'POST',
      body: {
        video_id: commentState.videoId,
        limit: 20,
        ...cursorBody(commentState)
      }
    });
    const comments = Array.isArray(data?.comments) ? data.comments : [];
    commentState.comments = reset ? comments : [...commentState.comments, ...comments];
    applyCursor(commentState, data);
  } catch (error) {
    commentState.error = error.message;
  } finally {
    commentState.loading = false;
  }
}

async function publishComment() {
  const content = commentState.draft.trim();
  if (!commentState.videoId) return;
  if (!isAuthed.value) {
    showToast('请先登录后评论', 'error');
    appPage.value = 'account';
    return;
  }
  if (!content) {
    commentState.error = '请输入评论内容';
    return;
  }

  commentState.publishing = true;
  commentState.error = '';
  try {
    const data = await apiRequest('/comment/publish', {
      method: 'POST',
      auth: true,
      body: {
        video_id: commentState.videoId,
        content
      }
    });
    if (data?.comment) {
      commentState.comments = [data.comment, ...commentState.comments];
    }
    updateVideoComments(commentState.videoId, data?.comments_count);
    commentState.draft = '';
    showToast('评论已发布');
  } catch (error) {
    commentState.error = error.message;
    showToast(error.message, 'error');
  } finally {
    commentState.publishing = false;
  }
}

async function deleteComment(comment) {
  const id = recordId(comment);
  if (!id || commentState.deletingById[id]) return;
  if (!isAuthed.value) {
    showToast('请先登录', 'error');
    appPage.value = 'account';
    return;
  }

  commentState.deletingById[id] = true;
  commentState.error = '';
  try {
    const data = await apiRequest('/comment/delete', {
      method: 'POST',
      auth: true,
      body: { comment_id: id }
    });
    commentState.comments = commentState.comments.filter((item) => recordId(item) !== id);
    updateVideoComments(commentState.videoId, data?.comments_count);
    showToast('评论已删除');
  } catch (error) {
    commentState.error = error.message;
    showToast(error.message, 'error');
  } finally {
    commentState.deletingById[id] = false;
  }
}

function canDeleteComment(comment) {
  return isAuthed.value && account.value?.id === comment?.account_id;
}

async function toggleLike(video) {
  const id = videoId(video);
  if (!id) return;
  if (!isAuthed.value) {
    showToast('请先登录后点赞', 'error');
    appPage.value = 'account';
    return;
  }
  if (likeLoadingById[id]) return;

  const currentlyLiked = isVideoLiked(video);
  likeLoadingById[id] = true;
  try {
    const data = await apiRequest(currentlyLiked ? '/like/unlike' : '/like/like', {
      method: 'POST',
      auth: true,
      body: { video_id: id }
    });
    setVideoLiked(id, !currentlyLiked);
    updateVideoLikes(id, data?.likes_count);
    if (currentlyLiked && currentView.value === 'liked') {
      likedFeed.videos = likedFeed.videos.filter((item) => videoId(item) !== id);
      if (activeVideoId.value === id) {
        setActiveVideo(likedFeed.videos[0] || latestFeed.videos[0]);
      }
    }
  } catch (error) {
    showToast(error.message, 'error');
  } finally {
    likeLoadingById[id] = false;
  }
}

function resetPlayer() {
  player.duration = 0;
  player.currentTime = 0;
  player.paused = true;
}

function syncPlayerState(video = activeVideoElement()) {
  if (!video) return;
  player.duration = Number.isFinite(video.duration) ? video.duration : 0;
  player.currentTime = Number.isFinite(video.currentTime) ? video.currentTime : 0;
  player.paused = video.paused;
  player.muted = video.muted;
}

function syncPlayerStateFor(video, event) {
  if (!isActiveVideo(video)) return;
  syncPlayerState(event?.currentTarget || activeVideoElement());
}

async function togglePlayback(video, event) {
  if (video) {
    setActiveVideo(video);
  }
  await nextTick();
  const mediaElement = event?.currentTarget?.tagName === 'VIDEO' ? event.currentTarget : activeVideoElement();
  if (!mediaElement) return;
  try {
    if (mediaElement.paused) {
      await mediaElement.play();
    } else {
      mediaElement.pause();
    }
    syncPlayerState(mediaElement);
  } catch (error) {
    showToast(error.message || '无法播放视频', 'error');
  }
}

function toggleMute(video) {
  if (video) {
    setActiveVideo(video);
  }
  const mediaElement = activeVideoElement();
  if (!mediaElement) return;
  mediaElement.muted = !mediaElement.muted;
  syncPlayerState(mediaElement);
}

function onSeekInput(event) {
  player.currentTime = Number(event.target.value || 0);
}

function commitSeek(event, video) {
  if (video) {
    setActiveVideo(video);
  }
  const mediaElement = activeVideoElement();
  if (!mediaElement) return;
  mediaElement.currentTime = Number(event.target.value || 0);
  syncPlayerState(mediaElement);
}

async function loadLatest(reset = false) {
  if (latestFeed.loading) return;
  latestFeed.loading = true;
  latestFeed.error = '';
  try {
    if (reset) {
      resetCursor(latestFeed);
      latestFeed.videos = [];
    }
    const data = await apiRequest('/feed/listLatest', {
      method: 'POST',
      body: {
        limit: 10,
        ...cursorBody(latestFeed)
      }
    });
    currentView.value = 'latest';
    appendFeed(latestFeed, data, reset, { updateActive: true });
    appPage.value = 'feed';
  } catch (error) {
    latestFeed.error = error.message;
  } finally {
    latestFeed.loading = false;
  }
}

async function loadByTag(reset = false, tagName = tagFeed.tag_name) {
  const normalizedTag = tagName.trim().replace(/^#/, '');
  if (!normalizedTag) {
    tagFeed.error = '请输入标签';
    return;
  }
  if (tagFeed.loading) return;
  tagFeed.loading = true;
  tagFeed.error = '';
  const shouldReset = reset || normalizedTag !== tagFeed.activeTag;
  try {
    if (shouldReset) {
      tagFeed.activeTag = normalizedTag;
      resetCursor(tagFeed);
      tagFeed.videos = [];
    }
    const data = await apiRequest('/feed/listByTag', {
      method: 'POST',
      body: {
        tag_name: normalizedTag,
        limit: 10,
        ...cursorBody(tagFeed)
      }
    });
    currentView.value = 'tag';
    appendFeed(tagFeed, data, shouldReset, { updateActive: true });
    appPage.value = 'feed';
  } catch (error) {
    tagFeed.error = error.message;
  } finally {
    tagFeed.loading = false;
  }
}

async function loadLikedVideos(reset = false) {
  if (!isAuthed.value) {
    showToast('请先登录后查看已赞视频', 'error');
    appPage.value = 'account';
    return;
  }
  if (likedFeed.loading) return;
  likedFeed.loading = true;
  likedFeed.error = '';
  try {
    if (reset) {
      resetCursor(likedFeed);
      likedFeed.videos = [];
    }
    const data = await apiRequest('/like/listLikedVideos', {
      method: 'POST',
      auth: true,
      body: {
        limit: 10,
        ...cursorBody(likedFeed)
      }
    });
    currentView.value = 'liked';
    appendFeed(likedFeed, data, reset, { updateActive: true });
    for (const video of likedFeed.videos) {
      setVideoLiked(videoId(video), true);
    }
    appPage.value = 'feed';
  } catch (error) {
    likedFeed.error = error.message;
  } finally {
    likedFeed.loading = false;
  }
}

async function loadAuthorVideos(authorId, reset = false) {
  if (!authorId || authorFeed.loading) return;
  authorFeed.loading = true;
  authorFeed.error = '';
  try {
    if (reset || authorFeed.author_id !== authorId) {
      authorFeed.author_id = authorId;
      resetCursor(authorFeed);
      authorFeed.videos = [];
    }
    await loadProfile(authorId, reset);
    const data = await apiRequest('/video/listByAuthorID', {
      method: 'POST',
      body: {
        author_id: authorId,
        limit: 10,
        ...cursorBody(authorFeed)
      }
    });
    appendFeed(authorFeed, data, reset);
  } catch (error) {
    authorFeed.error = error.message;
  } finally {
    authorFeed.loading = false;
  }
}

function appendFeed(feed, data, reset, options = {}) {
  const videos = Array.isArray(data?.videos) ? data.videos : Array.isArray(data?.likes) ? data.likes : [];
  feed.videos = reset ? videos : [...feed.videos, ...videos];
  applyCursor(feed, data);
  if (options.updateActive && videos.length && (!activeVideoId.value || reset)) {
    setActiveVideo(videos[0]);
  }
}

async function openDetail(video) {
  selectedVideo.value = video;
  detailLoading.value = true;
  try {
    const data = await apiRequest('/video/getDetail', {
      method: 'POST',
      body: { id: videoId(video) }
    });
    selectedVideo.value = data.video || video;
  } catch (error) {
    showToast(error.message, 'error');
  } finally {
    detailLoading.value = false;
  }
}

async function submitAuth() {
  if (!authForm.username.trim() || !authForm.password) {
    showToast('请输入用户名和密码', 'error');
    return;
  }
  authLoading.value = true;
  try {
    const path = authMode.value === 'register' ? '/account/register' : '/account/login';
    const data = await apiRequest(path, {
      method: 'POST',
      body: {
        username: authForm.username.trim(),
        password: authForm.password
      }
    });

    if (authMode.value === 'register') {
      showToast('注册成功，请登录');
      authMode.value = 'login';
      authForm.password = '';
    } else {
      token.value = data.token;
      account.value = data.account;
      localStorage.setItem(STORAGE_TOKEN, token.value);
      localStorage.setItem(STORAGE_ACCOUNT, JSON.stringify(account.value));
      showToast('登录成功');
      authForm.password = '';
      await loadProfile(account.value.id, true);
      await loadActiveLikeState();
      await loadActiveAuthorFollowState();
    }
  } catch (error) {
    showToast(error.message, 'error');
  } finally {
    authLoading.value = false;
  }
}

async function refreshMe(notify = true) {
  if (!token.value) return;
  meLoading.value = true;
  try {
    const data = await apiRequest('/account/me', {
      method: 'GET',
      auth: true
    });
    account.value = data;
    localStorage.setItem(STORAGE_ACCOUNT, JSON.stringify(data));
    await loadProfile(data.id, true);
    if (notify) showToast('个人信息已更新');
  } catch (error) {
    if (notify) showToast(error.message, 'error');
  } finally {
    meLoading.value = false;
  }
}

async function logout() {
  if (!token.value) return;
  try {
    await apiRequest('/account/logout', {
      method: 'POST',
      auth: true
    });
    showToast('已登出');
  } catch (error) {
    showToast(error.message, 'error');
  } finally {
    clearSession();
    appPage.value = 'account';
  }
}

function clearSession() {
  token.value = '';
  account.value = null;
  likedFeed.videos = [];
  resetCursor(likedFeed);
  commentState.draft = '';
  commentState.error = '';
  commentState.publishing = false;
  for (const id of Object.keys(commentState.deletingById)) {
    delete commentState.deletingById[id];
  }
  relationState.accounts = [];
  resetCursor(relationState);
  relationState.error = '';
  videoPanelOpen.value = videoPanelOpen.value && videoPanelTab.value !== 'messages';
  messageState.toId = null;
  messageState.title = '';
  messageState.messages = [];
  messageState.draft = '';
  resetCursor(messageState);
  messageState.loading = false;
  messageState.sending = false;
  messageState.error = '';
  for (const id of Object.keys(likedById)) {
    delete likedById[id];
  }
  for (const id of Object.keys(likeLoadingById)) {
    delete likeLoadingById[id];
  }
  for (const id of Object.keys(followingById)) {
    delete followingById[id];
  }
  for (const id of Object.keys(followLoadingById)) {
    delete followLoadingById[id];
  }
  localStorage.removeItem(STORAGE_TOKEN);
  localStorage.removeItem(STORAGE_ACCOUNT);
}

function onFileSelect(event, kind) {
  const file = event.target.files?.[0];
  event.target.value = '';
  if (!file) return;
  uploadState.error = '';
  if (kind === 'video') {
    uploadState.videoFile = file;
    publishForm.play_url = '';
    revokePreview(uploadState.videoPreview);
    uploadState.videoPreview = URL.createObjectURL(file);
  } else {
    uploadState.coverFile = file;
    publishForm.cover_url = '';
    revokePreview(uploadState.coverPreview);
    uploadState.coverPreview = URL.createObjectURL(file);
  }
}

function revokePreview(url) {
  if (url) URL.revokeObjectURL(url);
}

async function uploadFile(kind) {
  const isVideo = kind === 'video';
  const file = isVideo ? uploadState.videoFile : uploadState.coverFile;
  if (!isAuthed.value) {
    uploadState.error = '请先登录';
    appPage.value = 'account';
    return;
  }
  if (!file) {
    uploadState.error = isVideo ? '请选择视频文件' : '请选择封面图片';
    return;
  }

  const formData = new FormData();
  formData.append(isVideo ? 'video' : 'cover', file);
  uploadState[isVideo ? 'videoUploading' : 'coverUploading'] = true;
  uploadState.error = '';

  try {
    const data = await apiRequest(isVideo ? '/video/uploadVideo' : '/video/uploadCover', {
      method: 'POST',
      auth: true,
      form: true,
      body: formData
    });
    if (isVideo) {
      publishForm.play_url = data.play_url || '';
      showToast('视频上传完成');
    } else {
      publishForm.cover_url = data.cover_url || '';
      showToast('封面上传完成');
    }
  } catch (error) {
    uploadState.error = error.message;
    showToast(error.message, 'error');
  } finally {
    uploadState[isVideo ? 'videoUploading' : 'coverUploading'] = false;
  }
}

async function publishVideo() {
  if (!publishReady.value) {
    showToast(isAuthed.value ? '请填写标题并先上传视频文件' : '请先登录', 'error');
    if (!isAuthed.value) appPage.value = 'account';
    return;
  }

  uploadState.publishing = true;
  uploadState.error = '';
  const body = {
    title: publishForm.title.trim(),
    description: publishForm.description.trim(),
    play_url: publishForm.play_url.trim()
  };
  if (publishForm.cover_url.trim()) {
    body.cover_url = publishForm.cover_url.trim();
  }

  try {
    const data = await apiRequest('/video/publish', {
      method: 'POST',
      auth: true,
      body
    });
    showToast('发布成功');
    resetPublishForm();
    await loadLatest(true);
    if (data?.video) {
      setActiveVideo(data.video);
    }
  } catch (error) {
    uploadState.error = error.message;
    showToast(error.message, 'error');
  } finally {
    uploadState.publishing = false;
  }
}

function resetPublishForm() {
  publishForm.title = '';
  publishForm.description = '';
  publishForm.play_url = '';
  publishForm.cover_url = '';
  uploadState.videoFile = null;
  uploadState.coverFile = null;
  revokePreview(uploadState.videoPreview);
  revokePreview(uploadState.coverPreview);
  uploadState.videoPreview = '';
  uploadState.coverPreview = '';
}
</script>

<template>
  <main class="app-shell">
    <header class="topbar">
      <button class="brand" type="button" @click="loadLatest(true)">
        <span class="brand-mark">MF</span>
        <span>My Feed System</span>
      </button>

      <nav class="app-nav" aria-label="主导航">
        <button :class="{ active: appPage === 'feed' }" type="button" @click="goPage('feed')">浏览视频</button>
        <button :class="{ active: appPage === 'publish' }" type="button" @click="goPage('publish')">上传发布</button>
        <button :class="{ active: appPage === 'account' }" type="button" @click="goPage('account')">账户</button>
      </nav>

      <button class="session-pill" type="button" @click="goPage('account')">
        <span class="status-dot" :class="{ online: isAuthed }"></span>
        <span>{{ selectedAccountName }}</span>
      </button>
    </header>

    <section class="page-shell">
      <section v-if="appPage === 'feed'" class="feed-stage" aria-label="视频流">
        <div class="feed-toolbar">
          <div>
            <p class="eyebrow">
              {{ currentView === 'tag' ? 'Tag Feed' : currentView === 'liked' ? 'Liked Feed' : currentView === 'author' ? 'Author Feed' : 'Latest Feed' }}
            </p>
            <h1>{{ feedTitle }}</h1>
          </div>

          <div class="feed-actions">
            <div class="view-tabs" aria-label="内容视图">
              <button :class="{ active: currentView === 'latest' }" type="button" @click="loadLatest(true)">最新</button>
              <button :class="{ active: currentView === 'tag' }" type="button" @click="currentView = 'tag'">标签</button>
              <button :class="{ active: currentView === 'liked' }" type="button" @click="loadLikedVideos(true)">已赞</button>
            </div>
            <form class="tag-search" @submit.prevent="loadByTag(true)">
              <input v-model="tagFeed.tag_name" placeholder="搜索标签，如 go" />
              <button type="submit">搜索</button>
            </form>
          </div>
        </div>

        <div v-if="activeFeed.error" class="state-line error-text">{{ activeFeed.error }}</div>

        <div v-if="!visibleVideos.length && activeFeed.loading" class="empty-state">正在加载视频...</div>
        <div v-else-if="!visibleVideos.length" class="empty-state">
          <strong>还没有视频</strong>
          <span>启动后端后刷新，或先去上传发布一个视频。</span>
        </div>

        <div v-else class="feed-content" :class="{ 'panel-open': videoPanelOpen }">
          <div ref="reelScroller" class="reel-layout reel-scroll" aria-label="上下滑动视频流">
            <article
              v-for="video in visibleVideos"
              :key="videoId(video)"
              class="hero-video reel-item"
              :class="{ active: isActiveVideo(video) }"
              :data-reel-video-id="videoRefKey(video)"
            >
              <video
                v-if="video.play_url"
                :ref="(element) => setVideoRef(video, element)"
                :poster="normalizeUrl(video.cover_url)"
                :src="normalizeUrl(video.play_url)"
                loop
                playsinline
                preload="metadata"
                @click="togglePlayback(video, $event)"
                @durationchange="syncPlayerStateFor(video, $event)"
                @loadedmetadata="syncPlayerStateFor(video, $event)"
                @pause="syncPlayerStateFor(video, $event)"
                @play="syncPlayerStateFor(video, $event)"
                @timeupdate="syncPlayerStateFor(video, $event)"
              ></video>
              <img v-else-if="video.cover_url" :src="normalizeUrl(video.cover_url)" alt="" />
              <div v-else class="media-fallback">No Media</div>

              <div class="video-overlay">
                <div class="author-actions">
                  <button class="author-button" type="button" @click="openAuthorPanel(video.author_id)">
                    {{ profileName(video.author_id, `作者 ${video.author_id}`) }}
                  </button>
                  <button
                    v-if="!isOwnAccount(video.author_id)"
                    class="follow-button"
                    :class="{ following: isFollowing(video.author_id) }"
                    :disabled="isFollowLoading(video.author_id)"
                    type="button"
                    @click="toggleFollow(video.author_id)"
                  >
                    {{ isFollowing(video.author_id) ? '已关注' : '关注' }}
                  </button>
                  <button
                    v-if="!isOwnAccount(video.author_id)"
                    class="message-button"
                    type="button"
                    @click="openMessages(video.author_id, profileName(video.author_id, `作者 ${video.author_id}`))"
                  >
                    私信
                  </button>
                </div>
                <h2>{{ video.title }}</h2>
                <p>{{ video.description || '暂无描述' }}</p>
                <div class="tag-row">
                  <button v-for="tag in extractTags(video)" :key="tag" type="button" @click="loadByTag(true, tag)">
                    {{ tag }}
                  </button>
                </div>
              </div>

              <div class="metric-rail">
                <button
                  :class="{ liked: isVideoLiked(video) }"
                  :disabled="isLikeLoading(video)"
                  type="button"
                  :title="isVideoLiked(video) ? '取消点赞' : '点赞'"
                  @click="toggleLike(video)"
                >
                  <span>♥</span>
                  {{ compactNumber(video.likes_count) }}
                </button>
                <button type="button" title="评论" @click="openComments(video)">
                  <span>◎</span>
                  {{ compactNumber(video.comments_count) }}
                </button>
                <button type="button" title="热度">
                  <span>↗</span>
                  {{ compactNumber(video.popularity) }}
                </button>
                <button type="button" title="详情" @click="openDetail(video)">
                  <span>i</span>
                  详情
                </button>
              </div>

              <div v-if="isActiveVideo(video) && video.play_url" class="player-controls">
                <div class="player-actions">
                  <button type="button" @click="togglePlayback(video)">{{ player.paused ? '播放' : '暂停' }}</button>
                  <span>{{ formatDuration(player.currentTime) }} / {{ formatDuration(player.duration) }}</span>
                  <button type="button" @click="toggleMute(video)">{{ player.muted ? '取消静音' : '静音' }}</button>
                </div>
                <input
                  class="player-progress"
                  min="0"
                  step="0.1"
                  type="range"
                  :aria-valuemax="player.duration"
                  aria-label="视频进度"
                  :max="player.duration || 0"
                  :style="{ '--progress': progressPercent }"
                  :value="player.currentTime"
                  @change="commitSeek($event, video)"
                  @input="onSeekInput"
                />
              </div>
            </article>

            <button
              v-if="activeFeed.has_more"
              class="load-more reel-load-more"
              :disabled="activeFeed.loading"
              type="button"
              @click="loadMoreActiveFeed"
            >
              {{ activeFeed.loading ? '加载中...' : '加载更多' }}
            </button>
          </div>

          <aside v-if="videoPanelOpen" class="video-side-panel" aria-label="视频信息侧栏">
            <div class="drawer-head video-panel-head">
              <div>
                <p class="eyebrow">Panel</p>
                <h2>{{ videoPanelHeading }}</h2>
                <p class="drawer-subtitle">{{ videoPanelSubtitle }}</p>
              </div>
              <button type="button" @click="closeVideoPanel">×</button>
            </div>

            <div class="panel-tabs" aria-label="侧栏标签">
              <button
                v-for="tab in VIDEO_PANEL_TABS"
                :key="tab.id"
                :class="{ active: videoPanelTab === tab.id }"
                type="button"
                @click="switchVideoPanelTab(tab.id)"
              >
                {{ tab.label }}
              </button>
            </div>

            <div class="video-panel-body">
              <section v-if="videoPanelTab === 'author'" class="video-panel-section">
                <section class="profile-summary">
                  <div class="profile-head compact">
                    <div class="avatar">
                      <img v-if="profileAvatar(authorFeed.author_id)" :src="normalizeUrl(profileAvatar(authorFeed.author_id))" alt="" />
                      <span v-else>{{ profileName(authorFeed.author_id, 'U')?.slice(0, 1)?.toUpperCase() || 'U' }}</span>
                    </div>
                    <div>
                      <p class="profile-name">{{ profileName(authorFeed.author_id, `作者 ${authorFeed.author_id}`) }}</p>
                      <p class="muted">{{ profileBio(authorFeed.author_id, '这个人还没有填写简介。') || '这个人还没有填写简介。' }}</p>
                    </div>
                  </div>
                  <div v-if="authorProfile?.stats" class="profile-stats">
                    <div>
                      <strong>{{ compactNumber(authorProfile.stats.videos_count) }}</strong>
                      <span>作品</span>
                    </div>
                    <div>
                      <strong>{{ compactNumber(authorProfile.stats.likes_count) }}</strong>
                      <span>获赞</span>
                    </div>
                    <div>
                      <strong>{{ compactNumber(authorProfile.stats.followers_count) }}</strong>
                      <span>粉丝</span>
                    </div>
                    <div>
                      <strong>{{ compactNumber(authorProfile.stats.followings_count) }}</strong>
                      <span>关注</span>
                    </div>
                  </div>
                </section>

                <div class="drawer-actions">
                  <button
                    v-if="!isOwnAccount(authorFeed.author_id)"
                    class="primary-action"
                    :disabled="isFollowLoading(authorFeed.author_id)"
                    type="button"
                    @click="toggleFollow(authorFeed.author_id)"
                  >
                    {{ isFollowing(authorFeed.author_id) ? '已关注' : '关注作者' }}
                  </button>
                  <button class="ghost-action" type="button" @click="openRelations('followers', authorFeed.author_id, profileName(authorFeed.author_id, `作者 ${authorFeed.author_id}`))">粉丝</button>
                  <button class="ghost-action" type="button" @click="openRelations('following', authorFeed.author_id, profileName(authorFeed.author_id, `作者 ${authorFeed.author_id}`))">关注</button>
                  <button v-if="!isOwnAccount(authorFeed.author_id)" class="ghost-action" type="button" @click="openAuthorMessages(authorFeed.author_id)">私信</button>
                </div>

                <p v-if="authorFeed.error" class="error-text">{{ authorFeed.error }}</p>
                <div v-if="!authorFeed.videos.length && authorFeed.loading" class="empty-state slim">正在加载...</div>
                <div v-else-if="!authorFeed.videos.length" class="empty-state slim">暂无作品</div>

                <div class="drawer-list">
                  <button v-for="video in authorFeed.videos" :key="videoId(video)" type="button" @click="playPanelVideo(video)">
                    <img v-if="video.cover_url" :src="normalizeUrl(video.cover_url)" alt="" />
                    <span>
                      <strong>{{ video.title }}</strong>
                      <small>{{ formatTime(createdAtOf(video)) }}</small>
                    </span>
                  </button>
                </div>

                <button v-if="authorFeed.has_more" class="load-more" :disabled="authorFeed.loading" type="button" @click="loadAuthorVideos(authorFeed.author_id)">
                  {{ authorFeed.loading ? '加载中...' : '加载更多作品' }}
                </button>
              </section>

              <section v-else-if="videoPanelTab === 'comments'" class="video-panel-section">
                <form class="comment-form" @submit.prevent="publishComment">
                  <textarea v-model="commentState.draft" placeholder="写下你的评论"></textarea>
                  <button class="primary-action" :disabled="commentState.publishing || !commentState.draft.trim()" type="submit">
                    {{ commentState.publishing ? '发布中...' : '发布评论' }}
                  </button>
                </form>

                <p v-if="commentState.error" class="error-text">{{ commentState.error }}</p>
                <div v-if="!commentState.comments.length && commentState.loading" class="empty-state slim">正在加载...</div>
                <div v-else-if="!commentState.comments.length" class="empty-state slim">暂无评论</div>

                <div class="comment-list">
                  <article v-for="comment in commentState.comments" :key="recordId(comment)" class="comment-item">
                    <div class="comment-meta">
                      <span>用户 {{ comment.account_id }}</span>
                      <time>{{ formatTime(createdAtOf(comment)) }}</time>
                    </div>
                    <p>{{ comment.content }}</p>
                    <button
                      v-if="canDeleteComment(comment)"
                      class="comment-delete"
                      :disabled="commentState.deletingById[recordId(comment)]"
                      type="button"
                      @click="deleteComment(comment)"
                    >
                      {{ commentState.deletingById[recordId(comment)] ? '删除中' : '删除' }}
                    </button>
                  </article>
                </div>

                <button v-if="commentState.has_more" class="load-more" :disabled="commentState.loading" type="button" @click="loadComments(false)">
                  {{ commentState.loading ? '加载中...' : '加载更多评论' }}
                </button>
              </section>

              <section v-else class="video-panel-section message-panel-section">
                <p v-if="messageState.error" class="error-text">{{ messageState.error }}</p>
                <button v-if="messageState.has_more" class="load-more" :disabled="messageState.loading" type="button" @click="loadConversation(false)">
                  {{ messageState.loading ? '加载中...' : '加载更早消息' }}
                </button>

                <div v-if="!messageState.messages.length && messageState.loading" class="empty-state slim">正在加载...</div>
                <div v-else-if="!messageState.messages.length" class="empty-state slim">暂无私信</div>

                <div class="message-list">
                  <article
                    v-for="message in messageState.messages"
                    :key="recordId(message)"
                    class="message-item"
                    :class="{ mine: isOwnMessage(message) }"
                  >
                    <small>{{ isOwnMessage(message) ? '我' : messageState.title }} · {{ formatTime(createdAtOf(message)) }}</small>
                    <p>{{ message.content }}</p>
                  </article>
                </div>

                <form class="message-form" @submit.prevent="sendMessage">
                  <textarea v-model="messageState.draft" maxlength="1000" placeholder="输入私信内容"></textarea>
                  <button class="primary-action" :disabled="messageState.sending || !messageState.draft.trim() || !messageState.toId" type="submit">
                    {{ messageState.sending ? '发送中...' : '发送' }}
                  </button>
                </form>
              </section>
            </div>
          </aside>
        </div>
      </section>

      <section v-else-if="appPage === 'publish'" class="publish-page" aria-label="上传发布">
        <div class="page-heading">
          <p class="eyebrow">Publish</p>
          <h1>上传发布</h1>
        </div>

        <div v-if="!isAuthed" class="auth-lock">
          <strong>登录后才能上传视频</strong>
          <span>视频和封面会先上传到后端，后端返回地址后再创建视频记录。</span>
          <button class="primary-action" type="button" @click="goPage('account')">去登录</button>
        </div>

        <form v-else class="publish-workbench" @submit.prevent="publishVideo">
          <section class="panel-section upload-panel">
            <div class="section-heading">
              <p>文件上传</p>
              <span class="tiny-note">封面可选</span>
            </div>

            <div class="upload-grid">
              <label class="file-drop">
                <span>视频 .mp4</span>
                <input accept="video/mp4" type="file" @change="onFileSelect($event, 'video')" />
                <strong>{{ uploadState.videoFile?.name || '选择视频文件' }}</strong>
              </label>
              <label class="file-drop">
                <span>封面图片</span>
                <input accept=".jpg,.jpeg,.png,.webp,image/jpeg,image/png,image/webp" type="file" @change="onFileSelect($event, 'cover')" />
                <strong>{{ uploadState.coverFile?.name || '选择封面图片' }}</strong>
              </label>
            </div>

            <div class="button-row">
              <button class="ghost-action" :disabled="!uploadState.videoFile || uploadState.videoUploading" type="button" @click="uploadFile('video')">
                {{ uploadState.videoUploading ? '上传中' : hasUploadedVideo ? '重新上传视频' : '上传视频' }}
              </button>
              <button class="ghost-action" :disabled="!uploadState.coverFile || uploadState.coverUploading" type="button" @click="uploadFile('cover')">
                {{ uploadState.coverUploading ? '上传中' : hasUploadedCover ? '重新上传封面' : '上传封面' }}
              </button>
            </div>

            <div v-if="uploadState.videoPreview || uploadState.coverPreview" class="preview-strip">
              <video v-if="uploadState.videoPreview" :src="uploadState.videoPreview" muted playsinline controls></video>
              <img v-if="uploadState.coverPreview" :src="uploadState.coverPreview" alt="封面预览" />
            </div>

            <div class="upload-status-grid">
              <span :class="{ done: hasUploadedVideo }">{{ hasUploadedVideo ? '视频地址已自动填充' : '视频上传后才能发布' }}</span>
              <span :class="{ done: hasUploadedCover }">{{ hasUploadedCover ? '封面地址已自动填充' : '未上传封面时使用后端默认封面' }}</span>
            </div>
          </section>

          <section class="panel-section">
            <div class="section-heading">
              <p>发布信息</p>
              <span class="tiny-note">标题可包含 #tag</span>
            </div>

            <div class="stack-form">
              <label>
                <span>标题</span>
                <input v-model="publishForm.title" placeholder="First video #go #gin" />
              </label>
              <label>
                <span>描述</span>
                <textarea v-model="publishForm.description" placeholder="feed system #gorm"></textarea>
              </label>
            </div>

            <p v-if="uploadState.error" class="error-text">{{ uploadState.error }}</p>
            <button class="primary-action publish-submit" :disabled="!publishReady || uploadState.publishing" type="submit">
              {{ uploadState.publishing ? '发布中...' : '发布视频' }}
            </button>
          </section>
        </form>
      </section>

      <section v-else class="account-page" aria-label="账户">
        <div class="page-heading">
          <p class="eyebrow">Account</p>
          <h1>{{ isAuthed ? '个人信息' : '登录注册' }}</h1>
        </div>

        <section v-if="!isAuthed" class="panel-section account-card">
          <div class="section-heading">
            <p>{{ authMode === 'login' ? '欢迎回来' : '创建账号' }}</p>
            <div class="mode-switch">
              <button :class="{ active: authMode === 'login' }" type="button" @click="authMode = 'login'">登录</button>
              <button :class="{ active: authMode === 'register' }" type="button" @click="authMode = 'register'">注册</button>
            </div>
          </div>
          <form class="stack-form" @submit.prevent="submitAuth">
            <label>
              <span>用户名</span>
              <input v-model.trim="authForm.username" autocomplete="username" placeholder="user" />
            </label>
            <label>
              <span>密码</span>
              <input
                v-model="authForm.password"
                :autocomplete="authMode === 'login' ? 'current-password' : 'new-password'"
                placeholder="user"
                type="password"
              />
            </label>
            <button class="primary-action" :disabled="authLoading" type="submit">
              {{ authLoading ? '处理中...' : authMode === 'login' ? '登录' : '注册' }}
            </button>
          </form>
        </section>

        <section v-else class="panel-section account-card profile-section">
          <div class="profile-head">
            <div class="avatar">
              <img v-if="account?.avatar_url" :src="normalizeUrl(account.avatar_url)" alt="" />
              <span v-else>{{ account?.username?.slice(0, 1)?.toUpperCase() || 'U' }}</span>
            </div>
            <div>
              <p class="profile-name">{{ account?.username }}</p>
              <p class="muted">ID {{ account?.id }}</p>
            </div>
          </div>
          <p class="bio">{{ account?.bio || '这个人还没有填写简介。' }}</p>
          <div v-if="myProfile?.stats" class="profile-stats">
            <div>
              <strong>{{ compactNumber(myProfile.stats.videos_count) }}</strong>
              <span>作品</span>
            </div>
            <div>
              <strong>{{ compactNumber(myProfile.stats.likes_count) }}</strong>
              <span>获赞</span>
            </div>
            <div>
              <strong>{{ compactNumber(myProfile.stats.followers_count) }}</strong>
              <span>粉丝</span>
            </div>
            <div>
              <strong>{{ compactNumber(myProfile.stats.followings_count) }}</strong>
              <span>关注</span>
            </div>
          </div>
          <div class="button-row">
            <button class="ghost-action" :disabled="meLoading" type="button" @click="refreshMe()">
              {{ meLoading ? '刷新中' : '刷新资料' }}
            </button>
            <button class="ghost-action danger" type="button" @click="logout">登出</button>
          </div>
          <div class="button-row">
            <button class="ghost-action" type="button" @click="openRelations('followers', account.id, account.username)">我的粉丝</button>
            <button class="ghost-action" type="button" @click="openRelations('following', account.id, account.username)">我的关注</button>
          </div>
          <button class="secondary-action" type="button" @click="openAuthorPanel(account.id)">查看我的作品</button>
        </section>
      </section>
    </section>

    <aside v-if="relationDrawer" class="drawer relation-drawer" aria-label="关注关系">
      <div class="drawer-head">
        <div>
          <p class="eyebrow">Follow</p>
          <h2>{{ relationState.mode === 'followers' ? '粉丝列表' : '关注列表' }}</h2>
          <p class="drawer-subtitle">{{ relationState.title }}</p>
        </div>
        <button type="button" @click="relationDrawer = false">×</button>
      </div>

      <p v-if="relationState.error" class="error-text">{{ relationState.error }}</p>
      <div v-if="!relationState.accounts.length && relationState.loading" class="empty-state slim">正在加载...</div>
      <div v-else-if="!relationState.accounts.length" class="empty-state slim">暂无数据</div>

      <div class="relation-list">
        <article v-for="item in relationState.accounts" :key="recordId(item)" class="relation-item">
          <div class="avatar small">
            <img v-if="item.avatar_url" :src="normalizeUrl(item.avatar_url)" alt="" />
            <span v-else>{{ item.username?.slice(0, 1)?.toUpperCase() || 'U' }}</span>
          </div>
          <div class="relation-copy">
            <strong>{{ item.username || `用户 ${recordId(item)}` }}</strong>
            <small>{{ item.bio || '暂无简介' }}</small>
            <time>关注于 {{ formatTime(item.followed_at) }}</time>
          </div>
          <div v-if="!isOwnAccount(recordId(item))" class="relation-actions">
            <button
              class="ghost-action relation-follow"
              type="button"
              @click="openMessages(recordId(item), item.username || `用户 ${recordId(item)}`)"
            >
              私信
            </button>
            <button
              class="ghost-action relation-follow"
              :disabled="isFollowLoading(recordId(item))"
              type="button"
              @click="toggleFollow(recordId(item))"
            >
              {{ isFollowing(recordId(item)) ? '已关注' : '关注' }}
            </button>
          </div>
        </article>
      </div>

      <button v-if="relationState.has_more" class="load-more" :disabled="relationState.loading" type="button" @click="loadRelations(false)">
        {{ relationState.loading ? '加载中...' : '加载更多' }}
      </button>
    </aside>

    <div v-if="selectedVideo" class="modal-backdrop" @click.self="selectedVideo = null">
      <article class="detail-modal">
        <button class="close-button" type="button" @click="selectedVideo = null">×</button>
        <p class="eyebrow">Video Detail</p>
        <h2>{{ selectedVideo.title }}</h2>
        <p class="muted">ID {{ videoId(selectedVideo) }} · {{ formatTime(createdAtOf(selectedVideo)) }}</p>
        <div class="detail-media">
          <video
            v-if="selectedVideo.play_url"
            :poster="normalizeUrl(selectedVideo.cover_url)"
            :src="normalizeUrl(selectedVideo.play_url)"
            controls
            playsinline
          ></video>
        </div>
        <p>{{ selectedVideo.description || '暂无描述' }}</p>
        <dl class="detail-grid">
          <div>
            <dt>作者</dt>
            <dd>{{ selectedVideo.author_id }}</dd>
          </div>
          <div>
            <dt>点赞</dt>
            <dd>{{ selectedVideo.likes_count }}</dd>
          </div>
          <div>
            <dt>评论</dt>
            <dd>{{ selectedVideo.comments_count }}</dd>
          </div>
          <div>
            <dt>热度</dt>
            <dd>{{ selectedVideo.popularity }}</dd>
          </div>
        </dl>
        <p v-if="detailLoading" class="muted">正在同步详情...</p>
        <button class="secondary-action detail-comment-action" type="button" @click="openComments(selectedVideo)">
          查看评论
        </button>
      </article>
    </div>

    <div v-if="toast.text" class="toast" :class="toast.type">{{ toast.text }}</div>
  </main>
</template>
