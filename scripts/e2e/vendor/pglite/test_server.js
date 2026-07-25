const { PGlite } = require('@electric-sql/pglite');
const { createServer } = require('pglite-server');

async function main() {
  const db = new PGlite();
  const server = createServer(db);
  server.listen(54320, () => {
    console.log('Server started on port 54320');
    server.close();
  });
}
main();
