FROM node:18

WORKDIR /work
RUN npm install -g pnpm
VOLUME ["/work"]
EXPOSE 3000

CMD ["/bin/bash", "-c", "echo VITE_DATA_API_URL=http://127.0.0.1:8787 >.env && pnpm i && pnpm run dev --host 0.0.0.0 --port 3000"]
