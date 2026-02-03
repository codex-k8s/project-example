import axios from 'axios';

export const http = axios.create({
  baseURL: '/api/v1',
  withCredentials: true,
  timeout: 10_000
});

