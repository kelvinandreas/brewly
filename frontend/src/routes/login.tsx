import { createRoute, useNavigate } from '@tanstack/react-router'
import { Route as rootRoute } from './__root'
import { useEffect, useState } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { useAuth } from '../hooks/useAuth'
import { ApiError } from '../lib/api'

export const Route = createRoute({
  getParentRoute: () => rootRoute,
  path: '/login',
  component: LoginPage,
})

const loginSchema = z.object({
  email: z.string().email('Valid email required'),
  password: z.string().min(6, 'Password must be at least 6 characters'),
})

const registerSchema = z.object({
  name: z.string().min(1, 'Name is required'),
  email: z.string().email('Valid email required'),
  password: z.string().min(8, 'Password must be at least 8 characters'),
})

type LoginForm = z.infer<typeof loginSchema>
type RegisterForm = z.infer<typeof registerSchema>

function LoginPage() {
  const navigate = useNavigate()
  const { user, isOwnerNotExists, login, registerOwner } = useAuth()
  const [mode, setMode] = useState<'login' | 'register'>('login')
  const [serverError, setServerError] = useState<string | null>(null)

  useEffect(() => {
    if (user) navigate({ to: '/dashboard' as string })
  }, [user, navigate])

  useEffect(() => {
    if (isOwnerNotExists) setMode('register')
  }, [isOwnerNotExists])

  const loginForm = useForm<LoginForm>({ resolver: zodResolver(loginSchema) })
  const registerForm = useForm<RegisterForm>({ resolver: zodResolver(registerSchema) })

  async function handleLogin(data: LoginForm) {
    setServerError(null)
    try {
      await login.mutateAsync(data)
      navigate({ to: '/dashboard' as string })
    } catch (err) {
      setServerError(err instanceof ApiError ? err.message : 'Login failed')
    }
  }

  async function handleRegister(data: RegisterForm) {
    setServerError(null)
    try {
      await registerOwner.mutateAsync(data)
      navigate({ to: '/dashboard' as string })
    } catch (err) {
      setServerError(err instanceof ApiError ? err.message : 'Registration failed')
    }
  }

  if (mode === 'register') {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50 p-4">
        <div className="w-full max-w-md space-y-6">
          <div>
            <h1 className="text-2xl font-bold text-gray-900">Set up Brewly</h1>
            <p className="mt-1 text-sm text-gray-500">Create your owner account to get started.</p>
          </div>

          {serverError && (
            <div className="rounded-md bg-red-50 p-3 text-sm text-red-700">{serverError}</div>
          )}

          <form onSubmit={registerForm.handleSubmit(handleRegister)} className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700">Name</label>
              <input
                {...registerForm.register('name')}
                className="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
                placeholder="Your name"
              />
              {registerForm.formState.errors.name && (
                <p className="mt-1 text-xs text-red-600">{registerForm.formState.errors.name.message}</p>
              )}
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700">Email</label>
              <input
                {...registerForm.register('email')}
                type="email"
                className="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
                placeholder="owner@cafe.com"
              />
              {registerForm.formState.errors.email && (
                <p className="mt-1 text-xs text-red-600">{registerForm.formState.errors.email.message}</p>
              )}
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700">Password</label>
              <input
                {...registerForm.register('password')}
                type="password"
                className="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
                placeholder="Min. 8 characters"
              />
              {registerForm.formState.errors.password && (
                <p className="mt-1 text-xs text-red-600">{registerForm.formState.errors.password.message}</p>
              )}
            </div>
            <button
              type="submit"
              disabled={registerOwner.isPending}
              className="w-full rounded-md bg-indigo-600 px-4 py-2 text-sm font-semibold text-white hover:bg-indigo-700 disabled:opacity-60"
            >
              {registerOwner.isPending ? 'Creating account…' : 'Create owner account'}
            </button>
          </form>

          <p className="text-center text-sm text-gray-500">
            Already have an account?{' '}
            <button onClick={() => setMode('login')} className="text-indigo-600 hover:underline">
              Sign in
            </button>
          </p>
        </div>
      </div>
    )
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50 p-4">
      <div className="w-full max-w-md space-y-6">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Brewly</h1>
          <p className="mt-1 text-sm text-gray-500">Sign in to your account.</p>
        </div>

        {serverError && (
          <div className="rounded-md bg-red-50 p-3 text-sm text-red-700">{serverError}</div>
        )}

        <form onSubmit={loginForm.handleSubmit(handleLogin)} className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700">Email</label>
            <input
              {...loginForm.register('email')}
              type="email"
              className="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
              placeholder="you@cafe.com"
            />
            {loginForm.formState.errors.email && (
              <p className="mt-1 text-xs text-red-600">{loginForm.formState.errors.email.message}</p>
            )}
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700">Password</label>
            <input
              {...loginForm.register('password')}
              type="password"
              className="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
            />
            {loginForm.formState.errors.password && (
              <p className="mt-1 text-xs text-red-600">{loginForm.formState.errors.password.message}</p>
            )}
          </div>
          <button
            type="submit"
            disabled={login.isPending}
            className="w-full rounded-md bg-indigo-600 px-4 py-2 text-sm font-semibold text-white hover:bg-indigo-700 disabled:opacity-60"
          >
            {login.isPending ? 'Signing in…' : 'Sign in'}
          </button>
        </form>

        {!isOwnerNotExists && (
          <p className="text-center text-sm text-gray-500">
            First time?{' '}
            <button onClick={() => setMode('register')} className="text-indigo-600 hover:underline">
              Register as owner
            </button>
          </p>
        )}
      </div>
    </div>
  )
}
