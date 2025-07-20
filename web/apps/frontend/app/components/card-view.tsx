import { useSearchParams } from 'react-router';
import { type JSX, useEffect, useState } from 'react';
import { Button, Container, DropdownMenu, Flex, IconButton, ScrollArea } from '@radix-ui/themes';
import { ArrowLeftIcon, ArrowRightIcon } from '@radix-ui/react-icons';



export type SearchParams = {
    page: string,
    pageSize: string
}

export const defaultSearchParams = {
    page: "0",
    pageSize: "5"
}

function transformPageQuery(params: URLSearchParams) {
    const page = Number(params.get('page')) ?? 0;
    const pageSize = Number(params.get('pageSize')) ?? 5;


    return {
        page,
        pageSize
    }
}


function useCardViewSearchParams<U extends { id: string }>(
    {
        searchParams,
        queryFn,
    }: {
        searchParams?: URLSearchParams,
        queryFn: (search : URLSearchParams) => Promise<U[]>
    }

) {
    const initialSearchParams = searchParams ?? {};

    const [search, setSearch] = useSearchParams({
        ...defaultSearchParams,
        ...initialSearchParams,
    });
    const [entries, setEntries] = useState<U[]>([]);

    useEffect(() => {
        queryFn(search).then(setEntries);
    }, [search])

    function prevPage() {
        setSearch({
            page: (Number(search.get('page')) - 1).toString(),
            pageSize: (Number(search.get('pageSize'))).toString()
        })
    }

    function nextPage() {
        setSearch({
            page: (Number(search.get('page')) + 1).toString(),
            pageSize: (Number(search.get('pageSize'))).toString()
        })
    }

    function setPageSize(pageSize: number) {
        setSearch({
            page: (Number(search.get('page'))).toString(),
            pageSize: pageSize.toString()
        })
    }

    return{
        entries,
        search,
        setSearch,
        prevPage,
        nextPage,
        setPageSize
    }
}

export function CardView<T extends { id: string }>({  searchParams, queryFn, renderCard}: {  searchParams?: URLSearchParams, queryFn: (search: URLSearchParams) => Promise<T[]>, renderCard: (entry: T) => JSX.Element }) {
    const {entries, search, setSearch, prevPage, nextPage, setPageSize} = useCardViewSearchParams<T>({
        searchParams,
        queryFn
    })

    return (
        <Container>
            <Flex gap="3" align="center" justify={'center'} direction='column' width={'100%'}>
                <Flex gap="3" align="center" justify={'center'} direction='row' width={'100%'}>
                    <IconButton onClick={prevPage}>
                        <ArrowLeftIcon width="18" height="18" />
                    </IconButton>
                    <DropdownMenu.Root >
                        <DropdownMenu.Trigger>
                            <Button variant="soft" >
                                {search.get('pageSize')}
                                <DropdownMenu.TriggerIcon />
                            </Button>
                        </DropdownMenu.Trigger>
                        <DropdownMenu.Content >
                            {[5, 10, 20, 50, 100].map((size) => (
                                <DropdownMenu.Item
                                    key={size}
                                    onSelect={() => setPageSize(size)}
                                >
                                    {size}
                                </DropdownMenu.Item>
                            ))}
                        </DropdownMenu.Content>
                    </DropdownMenu.Root>
                    <IconButton onClick={nextPage}>
                        <ArrowRightIcon width="18" height="18" />
                    </IconButton>
                </Flex>
                <ScrollArea type="always" scrollbars="vertical" style={{ height: 700 }}>

                    <Flex gap="3" align="center" direction='column' width={'100%'}>
                        {
                            entries.length > 0 && entries.map((entry) => (
                              renderCard(entry)
                            ))
                        }
                    </Flex>
                </ScrollArea>
            </Flex>
        </Container>
    );

}