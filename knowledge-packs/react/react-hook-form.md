# React Hook Form Reference

## useForm + register + handleSubmit
```tsx
import { useForm } from 'react-hook-form'

interface LoginForm {
  email: string
  password: string
}

function LoginForm() {
  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<LoginForm>({
    defaultValues: { email: '', password: '' },
  })

  const onSubmit = async (data: LoginForm) => {
    await signIn(data.email, data.password)
  }

  return (
    <form onSubmit={handleSubmit(onSubmit)}>
      <input
        {...register('email', {
          required: '이메일을 입력하세요',
          pattern: { value: /^\S+@\S+\.\S+$/, message: '올바른 이메일 형식이 아닙니다' },
        })}
      />
      {errors.email && <p>{errors.email.message}</p>}

      <input
        type="password"
        {...register('password', {
          required: '비밀번호를 입력하세요',
          minLength: { value: 8, message: '8자 이상 입력하세요' },
        })}
      />
      {errors.password && <p>{errors.password.message}</p>}

      <button type="submit" disabled={isSubmitting}>
        {isSubmitting ? '로그인 중...' : '로그인'}
      </button>
    </form>
  )
}
```

## formState 주요 필드
```tsx
const {
  formState: {
    errors,        // 필드별 에러 객체
    isSubmitting,  // handleSubmit 실행 중
    isDirty,       // 초기값과 달라진 필드 존재
    isValid,       // 모든 유효성 통과 여부
    dirtyFields,   // 변경된 필드 목록
    touchedFields, // 포커스 후 벗어난 필드 목록
  },
} = useForm<FormData>({ mode: 'onChange' })
```

## watch, setValue, getValues, reset, trigger
```tsx
const { watch, setValue, getValues, reset, trigger } = useForm<FormData>()

// watch: 실시간 값 구독 (리렌더 유발)
const watchedEmail = watch('email')
const allValues = watch()                     // 전체 구독
const [email, name] = watch(['email', 'name']) // 다중 구독

// setValue: 프로그래매틱 값 설정
setValue('email', 'user@example.com', {
  shouldValidate: true, // 설정 후 유효성 검사
  shouldDirty: true,    // dirty 상태로 표시
})

// getValues: 리렌더 없이 현재 값 읽기
const currentEmail = getValues('email')
const allCurrentValues = getValues()

// reset: 폼 초기화
reset()                           // defaultValues로 초기화
reset({ email: 'new@email.com' }) // 특정 값으로 초기화

// trigger: 수동 유효성 검사 실행
await trigger('email')   // 특정 필드
await trigger()          // 전체 필드
```

## Controller (controlled 컴포넌트)
```tsx
import { useForm, Controller } from 'react-hook-form'
import Select from 'react-select'

interface FormData {
  category: { value: string; label: string } | null
}

function Form() {
  const { control, handleSubmit } = useForm<FormData>()

  return (
    <form onSubmit={handleSubmit(console.log)}>
      <Controller
        name="category"
        control={control}
        rules={{ required: '카테고리를 선택하세요' }}
        render={({ field, fieldState: { error } }) => (
          <>
            <Select
              {...field}
              options={[
                { value: 'tech', label: '기술' },
                { value: 'life', label: '라이프' },
              ]}
            />
            {error && <p>{error.message}</p>}
          </>
        )}
      />
      <button type="submit">제출</button>
    </form>
  )
}
```

## Zod resolver 연동
```tsx
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'

const schema = z.object({
  email: z.string().email('올바른 이메일 형식이 아닙니다'),
  password: z.string().min(8, '8자 이상 입력하세요'),
  age: z.number().min(18, '18세 이상이어야 합니다'),
})

type FormData = z.infer<typeof schema>

function RegisterForm() {
  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<FormData>({
    resolver: zodResolver(schema),
    defaultValues: { email: '', password: '', age: 0 },
  })

  return (
    <form onSubmit={handleSubmit(console.log)}>
      <input {...register('email')} />
      {errors.email && <p>{errors.email.message}</p>}
      <input type="password" {...register('password')} />
      {errors.password && <p>{errors.password.message}</p>}
      <button type="submit">가입</button>
    </form>
  )
}
```

## mode 옵션
```tsx
useForm({
  mode: 'onSubmit',   // 기본값: 제출 시에만 유효성 검사
  mode: 'onChange',   // 입력할 때마다 검사 (성능 주의)
  mode: 'onBlur',     // 포커스 벗어날 때 검사
  mode: 'onTouched',  // 최초 blur 이후부터 onChange로 전환
  mode: 'all',        // onChange + onBlur 모두
})
```

## 자주 쓰는 패턴

### 서버 에러 표시 (setError)
```tsx
const { setError } = useForm<FormData>()

const onSubmit = async (data: FormData) => {
  const result = await loginApi(data)
  if (!result.ok) {
    setError('email', { message: '이메일 또는 비밀번호가 틀렸습니다' })
    return
  }
  // 성공 처리
}
```

### 동적 필드 배열 (useFieldArray)
```tsx
import { useFieldArray } from 'react-hook-form'

const { fields, append, remove } = useFieldArray({
  control,
  name: 'items',
})

return (
  <>
    {fields.map((field, index) => (
      <div key={field.id}>
        <input {...register(`items.${index}.name`)} />
        <button type="button" onClick={() => remove(index)}>삭제</button>
      </div>
    ))}
    <button type="button" onClick={() => append({ name: '' })}>추가</button>
  </>
)
```
