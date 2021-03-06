import { createSlice, PayloadAction } from '@reduxjs/toolkit';
import axios, { CancelTokenSource } from 'axios';
import { gCancelToken } from '../..';
import { movies, search as axiosSearch } from '../../axios';
import { IComment, ITranslatedMovie } from '../../models/MovieInfo';
import { AppDispatch } from '../store';
import { CountriesKeys, GenresKeys } from './FilterSlice';
import { getToken } from './UserSlice';

export interface MoviesItems {
  search?: ITranslatedMovie[];
  popular?: ITranslatedMovie[];
  byName?: ITranslatedMovie[];
}

export type MovieItemsKeys = keyof MoviesItems;
export interface CommentsItems {
  comments: IComment[];
  id?: string;
}
export interface ErrorItem {
  error: string;
}

export interface UpdateMoviesItem {
  id: string;
  views?: number;
  isViewed?: boolean;
}

export interface CancelTokenItem {
  source: CancelTokenSource | null;
}

export interface MoviesState {
  loading: boolean;
  error: string | null;
  movies: ITranslatedMovie[];
  search: string[];
  popular: string[];
  byName: string[];
  isEndOfMovies: boolean;
  // cancelToken: CancelTokenSource | null;
}

export interface IMoviesResponse {
  data: ITranslatedMovie[] | null;
  error?: string;
  status: boolean;
}

export interface IFilter {
  limit?: number;
  offset?: number;
  letter?: string;
  search?: string;
  genres?: GenresKeys[];
  genre?: string;
  years?: [] | number[][];
  countries?: CountriesKeys[];
  id?: string | number;
}
interface ILoaderInfo {
  activePeers: number;
  loadedPercent: number;
}

const initialState = {
  loading: false,
  error: null,
  movies: [],
  search: [],
  popular: [],
  byName: [],
  isEndOfMovies: false,
  cancelToken: null,
} as MoviesState;

const MoviesSlice = createSlice({
  name: 'movies',
  initialState,
  reducers: {
    addMovies(state, { payload }: PayloadAction<MoviesItems>) {
      const add = (type: MovieItemsKeys) => {
        console.log(
          '[loadMovies] payload, moviesLength',
          payload,
          state.movies.length
        );
        if (!payload[type] || !payload[type]?.length) return;
        if (state.movies.length) {
          const newMovies = payload[type]?.filter(
            (movie) => !state.movies.find((el) => el.en.id === movie.en.id)
          );
          console.log('[loadMovies] newMovies', newMovies);
          if (newMovies?.length) {
            state.movies = state.movies.concat(newMovies);
          }
          // set unique movie ids
          const set = new Set(
            state[type].concat((payload[type] || []).map((el) => el.en.id))
          );
          state[type] = Array.from(set);
        } else {
          const arr = payload[type] || [];
          state.movies = arr;
          state[type] = arr.map((el) => el.en.id);
        }
      };

      state.loading = false;
      if (payload.search) {
        add('search');
      } else if (payload.byName) {
        add('byName');
      } else if (payload.popular) {
        add('popular');
      } else {
        throw new Error('[movies:addMovies] no movies in payload');
      }
    },
    setMovies(state, { payload }: PayloadAction<MoviesItems>) {
      const set = (type: MovieItemsKeys) => {
        if (!payload[type]) return;
        // state[type] = payload[type] || [];
        const newMovies = payload[type]?.filter(
          (movie) => !state.movies.find((el) => el.en.id === movie.en.id)
        );
        if (newMovies) {
          state.movies = state.movies.concat(newMovies);
          state[type] = payload[type]?.map((el) => el.en.id) || [];
        }
      };
      state.loading = false;
      if (payload.search) {
        set('search');
      } else if (payload.byName) {
        set('byName');
      } else if (payload.popular) {
        set('popular');
      } else {
        throw new Error('[movies:setMovies] no movies in payload');
      }
    },
    updateMovie(state, { payload }: PayloadAction<UpdateMoviesItem>) {
      const { id, views, isViewed } = payload;
      const movie = state.movies.find((movie) => movie.en.id === id);
      if (movie && views) movie.en.info.views = views;
      if (movie && isViewed) movie.en.isViewed = isViewed;
    },
    updateComments(state, { payload }: PayloadAction<CommentsItems>) {
      state.loading = false;
      if (!payload.id)
        throw new Error('[movies:updateComments] no id in payload');
      if (!payload.comments)
        throw new Error('[movies:updateComments] no comments in payload');
      const movie = state.movies.find((movie) => movie.en.id === payload.id);
      if (!movie) throw new Error('[movies:updateComments] no movie found');
      const newComments = payload.comments.filter((comment) => {
        return movie.en.info.comments
          ? !movie.en.info.comments.find(
              (el) => el.commentid === comment.commentid
            )
          : true;
      });
      console.log(
        '[updateComments] newComments',
        newComments,
        payload.comments
      );
      movie.en.info.comments = movie.en.info.comments
        ? [...movie.en.info.comments, ...newComments]
        : [...newComments];
    },
    setEndOfMovies(state) {
      state.isEndOfMovies = true;
    },
    resetEndOfMovies(state) {
      state.isEndOfMovies = false;
    },
    setError(state, { payload }: PayloadAction<ErrorItem>) {
      state.error = payload.error;
      state.loading = false;
    },
    resetError(state) {
      state.error = null;
      state.loading = false;
    },
    startLoading(state) {
      state.loading = true;
    },
    stopLoading(state) {
      state.loading = false;
    },
  },
});

const loadMoviesAsync = async (
  params?: IFilter,
  source?: CancelTokenSource
): Promise<IMoviesResponse | null> => {
  console.log('[loadMoviesAsync] params', params);
  let res = null;
  if (params?.search)
    res = await axiosSearch('find', {
      cancelToken: source?.token,
      params: {
        category: 'All',
        search: params?.search,
      },
      headers: {
        accessToken: getToken(),
      },
    });
  else if (params?.letter)
    res = await movies(`byname`, {
      cancelToken: source?.token,
      params,
      headers: {
        accessToken: getToken(),
      },
    });
  else if (params?.genre)
    res = await movies(`bygenre`, {
      cancelToken: source?.token,
      params,
      headers: {
        accessToken: getToken(),
      },
    });
  else
    res = await movies('movies', {
      cancelToken: source?.token,
      params,
      headers: {
        accessToken: getToken(),
      },
    });

  // console.log(res.data);
  if (res.status === 404) return { data: [], status: false };
  else if (res.status >= 400) return null;
  else return res.data;
};

export const loadMovies = ({
  filter = {
    limit: 20,
    offset: 0,
  },
  callback,
  source,
}: {
  filter?: IFilter;
  callback?: (arg0: boolean) => void;
  source?: CancelTokenSource;
}) => async (dispatch: AppDispatch) => {
  console.log('[loadMovies] start loading', new Date().getTime());
  dispatch(startLoading());
  const limit = filter.limit || 20;
  console.log('[loadMovies] filter', filter);
  let movies;
  console.log('[loadMovies] res', movies);

  try {
    movies = await loadMoviesAsync(filter, source);
  } catch (e) {
    if (axios.isCancel(e)) {
      console.log('[loadMovies] request is cancelled', new Date().getTime());
      dispatch(stopLoading());
      gCancelToken.source = null;
      return;
    }
    console.log(e);
    return null;
  }

  if (movies && !movies.status && movies.data?.length === 0) {
    // 404
    dispatch(stopLoading());
    dispatch(setError({ error: `Cannot find movies` }));
    return null;
  } else if (movies?.status && movies.data) {
    // any succedeed respose
    if (movies.data.length < limit) dispatch(setEndOfMovies());
    const removedNulls = movies.data?.map((el) =>
      !el.ru ? { en: el.en, ru: el.en } : el
    );
    console.log('[loadMovies] removedNulls:', removedNulls);
    if (filter.letter) {
      dispatch(addMovies({ byName: removedNulls }));
    } else if (filter.search || filter.genre) {
      filter.offset && filter.offset > 0
        ? dispatch(addMovies({ search: removedNulls }))
        : dispatch(setMovies({ search: removedNulls }));
    } else {
      dispatch(addMovies({ popular: removedNulls }));
    }
    if (callback) {
      callback(movies && movies.data ? movies.data.length < limit : true);
    }
    return removedNulls;
  } else {
    // all failed responses
    dispatch(stopLoading());
    dispatch(setError({ error: 'Loading error' }));
    return null;
  }
};

export const loadActivePeers = async (
  id: string
): Promise<ILoaderInfo | null> => {
  let data = null;
  try {
    const res = await axios(`/api/loader/stats/${id}`);
    if (res.status < 400 && res.data.status) {
      data = res.data.data as ILoaderInfo;
    }
  } catch (e) {
    console.log(e);
  }
  return data;
};

export const loadMovie = (id: string) => async (dispatch: AppDispatch) => {
  dispatch(startLoading());
  const movie = await loadMoviesAsync({ id });
  console.log('loadMovie', movie);
  if (movie && !movie.status && movie.data?.length === 0) {
    // 404
    dispatch(setError({ error: `Cannot load movie` }));
  } else if (movie?.status && movie.data) {
    // any succedeed respose
    dispatch(addMovies({ search: movie.data }));
  } else {
    // all failed responses
    dispatch(setError({ error: 'Loading error' }));
  }
};

export const updateViews = (movieId: string) => async (
  dispatch: AppDispatch
) => {
  const res = await axios.patch(
    '/api/movies/views',
    { movieId },
    {
      headers: {
        accessToken: getToken(),
      },
    }
  );
  console.log(res.data);
  if (res.data.status) {
    dispatch(updateMovie({ id: movieId, views: +res.data.data[0].views }));
  }
};

/**
 * loads comments
 * @param params object with values:
 *    @param movieId movieId to find it in redux
 *    @param limit number of comments to fetch, default 5
 *    @param offset comments offset, default 0
 * @param callback fires when end of comments is reached
 */
export const loadComments = (
  {
    movieId,
    limit = 5,
    offset = 0,
  }: { movieId: string; limit?: number; offset?: number },
  callback?: (length: number) => void
) => async (dispatch: AppDispatch) => {
  dispatch(startLoading());
  const res = await movies('comments', {
    params: {
      movieId,
      limit,
      offset,
    },
  });
  console.log('loadComments', res.data);
  if (res.data.status) {
    const comments = res.data.data as IComment[];
    comments.length && dispatch(updateComments({ comments, id: movieId }));
    if (callback) callback(comments.length);
  }
};

export const {
  addMovies,
  setMovies,
  updateMovie,
  updateComments,
  setEndOfMovies,
  resetEndOfMovies,
  setError,
  resetError,
  startLoading,
  stopLoading,
} = MoviesSlice.actions;

export default MoviesSlice.reducer;
