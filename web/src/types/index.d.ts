declare module '*.png' {
  const content: any
  export default content
}

declare module '*.jpg' {
  const content: any
  export default content
}

declare module '*.gif' {
  const content: any
  export default content
}

declare module 'evergreen-ui' {
  export function Autocomplete(props: any): JSX.Element
  export function AutocompleteItem(props: any): JSX.Element
  export function CheckCircleIcon(props: any): JSX.Element
  export function Badge(props: any): JSX.Element
  export function Button(props: any): JSX.Element
  export function DangerIcon(props: any): JSX.Element
  export function Dialog(props: any): JSX.Element
  export function Icon(props: any): JSX.Element
  export function IconButton(props: any): JSX.Element
  export function Label(props: any): JSX.Element
  export function Link(props: any): JSX.Element
  export function Manager(props: any): JSX.Element
  export function Pane(props: any): JSX.Element
  export function Popover(props: any): JSX.Element
  export function Text(props: any): JSX.Element
  export function TextInput(props: any): JSX.Element
  export function TextInputField(props: any): JSX.Element
  export function TriangleIcon(props: any): JSX.Element
  export function Tooltip(props: any): JSX.Element
  export function SegmentedControl(props: any): JSX.Element
  export function SelectMenu(props: any): JSX.Element
  export function Spinner(props: any): JSX.Element
  export function Strong(props: any): JSX.Element
  export const toaster: any
  export const Position: any
  export const colors: any
}

interface ICrawling {
  id: number
  url: string
  createdAt: string
  processed: boolean
  pageResults: IPageResult[]
}

interface IPageResult {
  url: string
  status: number
  crawlingId: number
}

type StoreAction<A, P> = {
  type: A
  payload: P
}

type CrawlingsState = {
  readonly [index: number]: ICrawling
}

type ModalState = {
  readonly show: boolean
  readonly modalType?: 'CreateCrawling' | 'DeleteCrawling'
  readonly modalParams?: any
  readonly loading: boolean
}

type StoreState = {
  readonly crawlings: CrawlingsState
  readonly modal: ModalState
}
