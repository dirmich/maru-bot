import { Elysia, t } from 'elysia';
import { cors } from '@elysiajs/cors';
import { jwt } from '@elysiajs/jwt';
import { initDb, query } from './db';
import dotenv from 'dotenv';
import axios from 'axios';

dotenv.config();

const app = new Elysia()
  .use(cors())
  .use(
    jwt({
      name: 'jwt',
      secret: process.env.JWT_SECRET || 'marubot-secret-123',
    })
  )
  .derive(async ({ jwt, cookie: { auth } }) => {
    const profile = auth?.value ? await jwt.verify(auth.value as string) : null;
    return {
      user: profile as any
    };
  })
  .get('/', () => ({ status: 'Marubot Admin Backend Running' }))
  
  // Auth Routes
  .group('/auth', (app) => 
    app
      .get('/google', ({ set }) => {
        const clientId = process.env.GOOGLE_CLIENT_ID;
        const redirectUri = `${process.env.BACKEND_URL || 'http://localhost:4000'}/auth/google/callback`;
        const scope = 'openid email profile';
        const url = `https://accounts.google.com/o/oauth2/v2/auth?client_id=${clientId}&redirect_uri=${redirectUri}&response_type=code&scope=${scope}`;
        
        set.redirect = url;
      })
      
      .get('/google/callback', async ({ query: { code }, jwt, cookie: { auth }, set }) => {
        if (!code) return { error: 'No code provided' };
        
        try {
          // 1. Exchange code for tokens
          const tokenRes = await axios.post('https://oauth2.googleapis.com/token', {
            code,
            client_id: process.env.GOOGLE_CLIENT_ID,
            client_secret: process.env.GOOGLE_CLIENT_SECRET,
            redirect_uri: `${process.env.BACKEND_URL || 'http://localhost:4000'}/auth/google/callback`,
            grant_type: 'authorization_code',
          });

          const { access_token } = tokenRes.data;

          // 2. Get user info from Google
          const userRes = await axios.get('https://www.googleapis.com/oauth2/v3/userinfo', {
            headers: { Authorization: `Bearer ${access_token}` },
          });

          const { email, name, picture: avatar_url } = userRes.data;
          const isSuperuser = email === process.env.SUPERUSER_EMAIL;

          // 3. Upsert user in DB
          const result = await query(
            `INSERT INTO users (email, name, avatar_url, is_superuser) 
             VALUES ($1, $2, $3, $4) 
             ON CONFLICT (email) DO UPDATE SET name = $2, avatar_url = $3, is_superuser = $4
             RETURNING id, email, name, avatar_url, is_superuser`,
            [email, name, avatar_url, isSuperuser]
          );

          const dbUser = result.rows[0];

          // 4. Generate JWT and set cookie
          if (auth) {
            auth.set({
              value: await jwt.sign(dbUser),
              httpOnly: true,
              maxAge: 7 * 86400,
              path: '/',
            });
          }

          // 5. Redirect to frontend
          set.redirect = process.env.FRONTEND_URL || 'http://localhost:3000';
        } catch (err) {
          console.error('Auth error:', err);
          return { error: 'Authentication failed' };
        }
      })

      .get('/me', ({ user }) => {
        if (!user) throw new Error('Unauthorized');
        return user;
      })
  )

  // Public/User Instance Reporting
  .post('/instances/report', async ({ body }) => {
    try {
      await query(
        `INSERT INTO instances (user_id, device_name, os, memory, storage, language, version, last_check_in)
         VALUES ($1, $2, $3, $4, $5, $6, $7, CURRENT_TIMESTAMP)
         ON CONFLICT (id) DO UPDATE SET 
           os = $3, memory = $4, storage = $5, language = $6, version = $7, last_check_in = CURRENT_TIMESTAMP`,
        [body.user_id, body.device_name, body.os, body.memory, body.storage, body.language, body.version]
      );
      return { success: true };
    } catch (err) {
      console.error('Report error:', err);
      return { error: 'Reporting failed' };
    }
  }, {
    body: t.Object({
      user_id: t.String(),
      device_name: t.String(),
      os: t.String(),
      memory: t.Number(),
      storage: t.Number(),
      language: t.String(),
      version: t.String()
    })
  })

  // Admin Protected Routes
  .group('/admin', (app) =>
    app
      .onBeforeHandle(({ user, set }) => {
        if (!user || !user.is_superuser) {
          set.status = 403;
          return { error: 'Forbidden' };
        }
      })
      .get('/stats', async () => {
        const usersCount = await query('SELECT count(*) FROM users');
        const instancesCount = await query('SELECT count(*) FROM instances');
        const osStats = await query('SELECT os, count(*) FROM instances GROUP BY os');
        
        return {
          total_users: parseInt(usersCount.rows[0].count),
          total_instances: parseInt(instancesCount.rows[0].count),
          platform_stats: osStats.rows.reduce((acc: any, row: any) => {
            acc[row.os.toLowerCase()] = parseInt(row.count);
            return acc;
          }, {})
        };
      })
      .get('/users', async () => {
        const result = await query(`
          SELECT u.*, 
            (SELECT json_agg(i) FROM instances i WHERE i.user_id = u.id) as instances
          FROM users u
          ORDER BY u.created_at DESC
        `);
        return result.rows;
      })
  )

  .listen(process.env.PORT || 4000);

initDb();

console.log(`🚀 Admin Backend is running at ${app.server?.hostname}:${app.server?.port}`);
