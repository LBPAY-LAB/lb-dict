Especificação do capitulo 19 do manual do Dict Bacen (versão 8.0)


19 Consulta de baldes 

O DICT possui um endpoint que permite aos participantes consultarem os seus limites de requisição à API do DICT e o tamanho dos seus baldes no momento da consulta. Esse endpoint tem como objetivo permitir que os participantes façam uma melhor gestão sobre seus baldes, no intuito de evitar seu esvaziamento. Os participantes podem realizar dois tipos de requisição: 
    a. listagem de políticas de limitação (policies_list): possibilita a obtenção de uma lista contendo, para cada política de requisições, a quantidade de fichas disponíveis no balde no momento da consulta, a capacidade máxima do balde, a quantidade de reposição periódica de fichas, o período de reposição de fichas e a categoria em que o participante está enquadrado (para o balde de consulta de chaves); 
    e b. consulta de política de limitação (policies_read): permite ao participante consultar, para uma política de requisições individual, a quantidade de fichas disponíveis no balde no momento da consulta, a capacidade máxima do balde, a quantidade de reposição periódica de fichas, o período de reposição de fichas e a categoria em que o participante está enquadrado (para o balde de consulta de chaves). 
    
Na consulta de política de limitação, o participante deve informar como input o nome da política de requisições


Tem alguns topicos da API do Dict que colocamos aqui como referência para o escopo do projecto:


Política de Limitação
O controle do fluxo de acesso aos serviços do DICT é realizado por meio de políticas de limitação de requisições, utilizando o algoritmo de token bucket. Para cada política de limitação, existe um balde com configurações específicas, de acordo com a seção limitação de requisições.

É possível, a qualquer momento, consultar da lista de políticas de limitação e o estado dos baldes.

Listar Políticas
Obtém a lista de políticas de limitação de acesso ao DICT para o participante requisitante.

Além da lista de políticas, o serviço informa também a categoria em que o participante se encontra, de acordo com a tabela presente no item ENTRIES_READ_PARTICIPANT_ANTISCAN da seção limitação de requisições.

Cada item da lista detém informações do estado do balde correspondente no momento da consulta.

header Parameters
PI-RequestingParticipant
required
string^[0-9]{8}$
Example: 12345678
Identificador SPB do participante (direto ou indireto) que faz a requisição.

Responses
200
OK

403
Forbidden

503
Service Unavailable


get
/policies/


Response samples
200403503
Content type
application/xml

Copy
<?xml version="1.0" encoding="UTF-8" ?>
<ListPoliciesResponse>
    <Signature></Signature>
    <CorrelationId>B20220902101925801123456781A6998</CorrelationId>
    <ResponseTime>2022-09-02T13:19:25.833Z</ResponseTime>
    <Category>A</Category>
    <Policies>
        <Policy>
            <AvailableTokens>0</AvailableTokens>
            <Capacity>0</Capacity>
            <RefillTokens>0</RefillTokens>
            <RefillPeriodSec>0</RefillPeriodSec>
            <Name>ENTRIES_READ_PARTICIPANT_ANTISCAN</Name>
        </Policy>
    </Policies>
</ListPoliciesResponse>
Consultar Política
Obtém o estado atual do balde do participante para a política informada.

path Parameters
Policy
required
string
Example: ENTRIES_READ_PARTICIPANT_ANTISCAN
header Parameters
PI-RequestingParticipant
required
string^[0-9]{8}$
Example: 12345678
Identificador SPB do participante (direto ou indireto) que faz a requisição.

Responses
200
OK

403
Forbidden

503
Service Unavailable


get
/policies/{policy}


Response samples
200403503
Content type
application/xml

Copy
<?xml version="1.0" encoding="UTF-8" ?>
<GetPolicyResponse>
    <Signature></Signature>
    <CorrelationId>B20220902102129115123456788B2B22</CorrelationId>
    <ResponseTime>2022-09-02T13:21:29.110Z</ResponseTime>
    <Category>A</Category>
    <Policy>
        <AvailableTokens>0</AvailableTokens>
        <Capacity>0</Capacity>
        <RefillTokens>0</RefillTokens>
        <RefillPeriodSec>0</RefillPeriodSec>
        <Name>ENTRIES_READ_PARTICIPANT_ANTISCAN</Name>
    </Policy>
</GetPolicyResponse>




Listar Políticas
Obtém a lista de políticas de limitação de acesso ao DICT para o participante requisitante.

Além da lista de políticas, o serviço informa também a categoria em que o participante se encontra, de acordo com a tabela presente no item ENTRIES_READ_PARTICIPANT_ANTISCAN da seção limitação de requisições.

Cada item da lista detém informações do estado do balde correspondente no momento da consulta.

header Parameters
PI-RequestingParticipant
required
string^[0-9]{8}$
Example: 12345678
Identificador SPB do participante (direto ou indireto) que faz a requisição.

Responses
200
OK

403
Forbidden

503
Service Unavailable


get
/policies/


Response samples
200403503
Content type
application/xml

Copy
<?xml version="1.0" encoding="UTF-8" ?>
<ListPoliciesResponse>
    <Signature></Signature>
    <CorrelationId>B20220902101925801123456781A6998</CorrelationId>
    <ResponseTime>2022-09-02T13:19:25.833Z</ResponseTime>
    <Category>A</Category>
    <Policies>
        <Policy>
            <AvailableTokens>0</AvailableTokens>
            <Capacity>0</Capacity>
            <RefillTokens>0</RefillTokens>
            <RefillPeriodSec>0</RefillPeriodSec>
            <Name>ENTRIES_READ_PARTICIPANT_ANTISCAN</Name>
        </Policy>
    </Policies>
</ListPoliciesResponse>
Consultar Política
Obtém o estado atual do balde do participante para a política informada.

path Parameters
Policy
required
string
Example: ENTRIES_READ_PARTICIPANT_ANTISCAN
header Parameters
PI-RequestingParticipant
required
string^[0-9]{8}$
Example: 12345678
Identificador SPB do participante (direto ou indireto) que faz a requisição.

Responses
200
OK

403
Forbidden

503
Service Unavailable


get
/policies/{policy}


Response samples
200403503
Content type
application/xml

Copy
<?xml version="1.0" encoding="UTF-8" ?>
<GetPolicyResponse>
    <Signature></Signature>
    <CorrelationId>B20220902102129115123456788B2B22</CorrelationId>
    <ResponseTime>2022-09-02T13:21:29.110Z</ResponseTime>
    <Category>A</Category>
    <Policy>
        <AvailableTokens>0</AvailableTokens>
        <Capacity>0</Capacity>
        <RefillTokens>0</RefillTokens>
        <RefillPeriodSec>0</RefillPeriodSec>
        <Name>ENTRIES_READ_PARTICIPANT_ANTISCAN</Name>
    </Policy>
</GetPolicyResponse>




Consultar Política
Obtém o estado atual do balde do participante para a política informada.

path Parameters
Policy
required
string
Example: ENTRIES_READ_PARTICIPANT_ANTISCAN
header Parameters
PI-RequestingParticipant
required
string^[0-9]{8}$
Example: 12345678
Identificador SPB do participante (direto ou indireto) que faz a requisição.

Responses
200
OK

403
Forbidden

503
Service Unavailable


get
/policies/{policy}


Response samples
200403503
Content type
application/xml

Copy
<?xml version="1.0" encoding="UTF-8" ?>
<GetPolicyResponse>
    <Signature></Signature>
    <CorrelationId>B20220902102129115123456788B2B22</CorrelationId>
    <ResponseTime>2022-09-02T13:21:29.110Z</ResponseTime>
    <Category>A</Category>
    <Policy>
        <AvailableTokens>0</AvailableTokens>
        <Capacity>0</Capacity>
        <RefillTokens>0</RefillTokens>
        <RefillPeriodSec>0</RefillPeriodSec>
        <Name>ENTRIES_READ_PARTICIPANT_ANTISCAN</Name>
    </Policy>
</GetPolicyResponse>




Limitação de requisições
Para preservar a estabilidade do serviço, as operações da API do DICT estão sujeitas a políticas de limitação de requisições. Especificamente para a operação de consulta de vínculo, há também limitação de requisições com a finalidade de prevenir ataques de varredura de dados. O algoritmo usado para implementar as políticas de limitação é o token bucket.

Uma política de limitação tem associado a ela um escopo, que pode ser o usuário final ou o participante. Cada política possui uma taxa de reposição de "fichas", um tamanho de "balde" e uma regra de contagem. A tabela abaixo define as políticas aplicáveis a cada operação da API.

Política	Escopo	Operações	Taxa de reposição	Tamanho do balde
ENTRIES_READ_USER_ANTISCAN	USER	getEntry	(*)	(*)
ENTRIES_READ_USER_ANTISCAN_V2	USER	getEntry	(*)	(*)
ENTRIES_READ_PARTICIPANT_ANTISCAN	PSP	getEntry	(**)	(**)
ENTRIES_STATISTICS_READ	PSP	getEntryStatistics	(**)	(**)
ENTRIES_WRITE	PSP	createEntry, deleteEntry	1200/min	36000
ENTRIES_UPDATE	PSP	updateEntry	600/min	600
CLAIMS_READ	PSP	getClaim	600/min	18000
CLAIMS_WRITE	PSP	createClaim, acknowledgeClaim, cancelClaim, confirmClaim, completeClaim	1200/min	36000
CLAIMS_LIST_WITH_ROLE	PSP	listClaims	40/min	200
CLAIMS_LIST_WITHOUT_ROLE	PSP	listClaims	10/min	50
SYNC_VERIFICATIONS_WRITE	PSP	createSyncVerification	10/min	50
CIDS_FILES_WRITE	PSP	createCidSetFile	40/dia	200
CIDS_FILES_READ	PSP	getCidSetFile	10/min	50
CIDS_EVENTS_LIST	PSP	listCidSetEvents	20/min	100
CIDS_ENTRIES_READ	PSP	getEntryByCid	1200/min	36000
INFRACTION_REPORTS_READ	PSP	getInfractionReport	600/min	18000
INFRACTION_REPORTS_WRITE	PSP	createInfractionReport, acknowledgeInfractionReport, cancelInfractionReport, closeInfractionReport	1200/min	36000
INFRACTION_REPORTS_LIST_WITH_ROLE	PSP	listInfractionReports	40/min	200
INFRACTION_REPORTS_LIST_WITHOUT_ROLE	PSP	listInfractionReports	10/min	50
KEYS_CHECK	PSP	checkKeys	70/min	70
REFUNDS_READ	PSP	getRefund	1200/min	36000
REFUNDS_WRITE	PSP	createRefund, cancelRefund, closeRefund	2400/min	72000
REFUND_LIST_WITH_ROLE	PSP	listRefunds	40/min	200
REFUND_LIST_WITHOUT_ROLE	PSP	listRefunds	10/min	50
FRAUD_MARKERS_READ	PSP	getFraudMarker	600/min	18000
FRAUD_MARKERS_WRITE	PSP	createFraudMarker, cancelFraudMarker	1200/min	36000
FRAUD_MARKERS_LIST	PSP	listFrauds	600/min	18000
PERSONS_STATISTICS_READ	PSP	getPersonStatistics	12000/min	36000
POLICIES_READ	PSP	getBucketState	60/min	200
POLICIES_LIST	PSP	listBucketStates	6/min	20
Regras de contagem das políticas
ENTRIES_READ_USER_ANTISCAN

Aplicável somente para chaves do tipo EMAIL e PHONE
status 200: subtrai 1
status 404: subtrai 20
ordem de pagamento enviada
adiciona 1 para categoria Pessoa Física (PF)
adiciona 2 para categoria Pessoa Jurídica (PJ)
ENTRIES_READ_USER_ANTISCAN_V2

Aplicável somente para chaves do tipo CPF, CNPJ e EVP
status 200: subtrai 1
status 404: subtrai 20
ordem de pagamento enviada
adiciona 1 para categoria Pessoa Física (PF)
adiciona 2 para categoria Pessoa Jurídica (PJ)
(*) O tamanho do balde das políticas ENTRIES_READ_USER_ANTISCAN e ENTRIES_READ_USER_ANTISCAN_V2 é categorizado de acordo com o tipo de usuário final realizando a consulta, Pessoa Física (PF) ou Pessoa Jurídica (PJ):

Categoria	Taxa de reposição	Tamanho do balde
PF	2/min	100
PJ	20/min	1.000
ENTRIES_READ_PARTICIPANT_ANTISCAN

Aplicável para todos os tipos de chaves (EMAIL, PHONE, CPF, CNPJ e EVP)
status 200: subtrai 1
status 404: subtrai 3
ordem de pagamento enviada: adiciona 1
(**) O tamanho do balde nesta política depende da categoria em que se enquadra do participante. As seguintes categorias são possíveis:

Categoria	Taxa de reposição	Tamanho do balde
A	25.000/min	50.000
B	20.000/min	40.000
C	15.000/min	30.000
D	8.000/min	16.000
E	2.500/min	5.000
F	250/min	500
G	25/min	250
H	2/min	50
CLAIMS_LIST_WITH_ROLE

Aplicável quando há filtragem por doador/reivindicador
status diferente de 500: subtrai 1
CLAIMS_LIST_WITHOUT_ROLE

Aplicável quando não há filtragem por doador/reivindicador
status diferente de 500: subtrai 1
REFUND_LIST_WITH_ROLE

Aplicável quando há filtragem por requisitante/contestado
status diferente de 500: subtrai 1
REFUND_LIST_WITHOUT_ROLE

Aplicável quando não há filtragem por requisitante/contestado
status diferente de 500: subtrai 1
Demais políticas

status diferente de 500: subtrai 1
Violação de limites
Caso o limite de uma política seja excedido, o que acontece quando a quantidade de fichas chega a zero, será retornada uma resposta de erro com status 429.


Token bucket

Article
Talk
Read
Edit
View history

Tools
From Wikipedia, the free encyclopedia
The token bucket is an algorithm used in packet-switched and telecommunications networks. It can be used to check that data transmissions, in the form of packets, conform to defined limits on bandwidth and burstiness (a measure of the unevenness or variations in the traffic flow). It can also be used as a scheduling algorithm to determine the timing of transmissions that will comply with the limits set for the bandwidth and burstiness: see network scheduler.

Overview
The token bucket algorithm is based on an analogy of a fixed capacity bucket into which tokens, normally representing a unit of bytes or a single packet of predetermined size, are added at a fixed rate. When a packet is to be checked for conformance to the defined limits, the bucket is inspected to see if it contains sufficient tokens at that time. If so, the appropriate number of tokens, e.g. equivalent to the length of the packet in bytes, are removed ("cashed in"), and the packet is passed, e.g., for transmission. The packet does not conform if there are insufficient tokens in the bucket, and the contents of the bucket are not changed. Non-conformant packets can be treated in various ways:

They may be dropped.
They may be enqueued for subsequent transmission when sufficient tokens have accumulated in the bucket.
They may be transmitted, but marked as being non-conformant, possibly to be dropped subsequently if the network is overloaded.
A conforming flow can thus contain traffic with an average rate up to the rate at which tokens are added to the bucket, and have a burstiness determined by the depth of the bucket. This burstiness may be expressed in terms of either a jitter tolerance, i.e. how much sooner a packet might conform (e.g. arrive or be transmitted) than would be expected from the limit on the average rate, or a burst tolerance or maximum burst size, i.e. how much more than the average level of traffic might conform in some finite period.

Algorithm
The token bucket algorithm can be conceptually understood as follows:

A token is added to the bucket every 
1
/
r
{\displaystyle 1/r} seconds.
The bucket can hold at the most 
b
{\displaystyle b} tokens. If a token arrives when the bucket is full, it is discarded.
When a packet (network layer PDU) of n bytes arrives,
if at least n tokens are in the bucket, n tokens are removed from the bucket, and the packet is sent to the network.
if fewer than n tokens are available, no tokens are removed from the bucket, and the packet is considered to be non-conformant.
Variations
Implementers of this algorithm on platforms lacking the clock resolution necessary to add a single token to the bucket every 
1
/
r
{\displaystyle 1/r} seconds may want to consider an alternative formulation. Given the ability to update the token bucket every S milliseconds, the number of tokens to add every S milliseconds = 
(
r
∗
S
)
/
1000
{\displaystyle (r*S)/1000}.

Properties
Average rate
Over the long run the output of conformant packets is limited by the token rate, 
r
{\displaystyle r}.

Burst size
Let 
M
{\displaystyle M} be the maximum possible transmission rate in bytes/second.

Then 
T
max
=
{
b
/
(
M
−
r
)
 if 
r
<
M
∞
 otherwise 
{\displaystyle T_{\text{max}}={\begin{cases}b/(M-r)&{\text{ if }}r<M\\\infty &{\text{ otherwise }}\end{cases}}} is the maximum burst time, that is the time for which the rate 
M
{\displaystyle M} is fully utilized.

The maximum burst size is thus 
B
max
=
T
max
∗
M
{\displaystyle B_{\text{max}}=T_{\text{max}}*M}

Uses
The token bucket can be used in either traffic shaping or traffic policing. In traffic policing, nonconforming packets may be discarded (dropped) or may be reduced in priority (for downstream traffic management functions to drop if there is congestion). In traffic shaping, packets are delayed until they conform. Traffic policing and traffic shaping are commonly used to protect the network against excess or excessively bursty traffic, see bandwidth management and congestion avoidance. Traffic shaping is commonly used in the network interfaces in hosts to prevent transmissions being discarded by traffic management functions in the network.

The token bucket algorithm is also used in controlling database IO flow.[1] In it, limitation applies to neither IOPS nor the bandwidth but rather to a linear combination of both. By defining tokens to be the normalized sum of IO request weight and its length, the algorithm makes sure that the time derivative of the aforementioned function stays below the needed threshold.

Comparison to leaky bucket
The token bucket algorithm is directly comparable to one of the two versions of the leaky bucket algorithm described in the literature.[2][3][4][5] This comparable version of the leaky bucket is described on the relevant Wikipedia page as the leaky bucket algorithm as a meter. This is a mirror image of the token bucket, in that conforming packets add fluid, equivalent to the tokens removed by a conforming packet in the token bucket algorithm, to a finite capacity bucket, from which this fluid then drains away at a constant rate, equivalent to the process in which tokens are added at a fixed rate.

There is, however, another version of the leaky bucket algorithm,[3] described on the relevant Wikipedia page as the leaky bucket algorithm as a queue. This is a special case of the leaky bucket as a meter, which can be described by the conforming packets passing through the bucket. The leaky bucket as a queue is therefore applicable only to traffic shaping, and does not, in general, allow the output packet stream to be bursty, i.e. it is jitter free. It is therefore significantly different from the token bucket algorithm.

These two versions of the leaky bucket algorithm have both been described in the literature under the same name. This has led to considerable confusion over the properties of that algorithm and its comparison with the token bucket algorithm. However, fundamentally, the two algorithms are the same, and will, if implemented correctly and given the same parameters, see exactly the same packets as conforming and nonconforming.

Hierarchical token bucket
The hierarchical token bucket (HTB) is a faster replacement for the class-based queueing (CBQ) queuing discipline in Linux.[6] It is useful for limiting each client's download/upload rate so that the limited client cannot saturate the total bandwidth.


Three clients sharing the same outbound bandwidth.

Conceptually, HTB is an arbitrary number of token buckets arranged in a hierarchy. The primary egress queuing discipline (qdisc) on any device is known as the root qdisc. The root qdisc will contain one class. This single HTB class will be set with two parameters, a rate and a ceil. These values should be the same for the top-level class, and will represent the total available bandwidth on the link.

In HTB, rate means the guaranteed bandwidth available for a given class and ceil (short for ceiling) indicates the maximum bandwidth that class is allowed to consume. When a class requests a bandwidth more than guaranteed, it may borrow bandwidth from its parent as long as both ceils are not reached. Hierarchical Token Bucket implements a classful queuing mechanism for the Linux traffic control system, and provides rate and ceil to allow the user to control the absolute bandwidth to particular classes of traffic as well as indicate the ratio of distribution of bandwidth when extra bandwidth become available (up to ceil).

When choosing the bandwidth for a top-level class, traffic shaping only helps at the bottleneck between the LAN and the Internet. Typically, this is the case in home and office network environments, where an entire LAN is serviced by a DSL or T1 connection.

See also
Rate limiting
Traffic shaping
Counting semaphores
References
 "Implementing a New IO Scheduler Algorithm for Mixed Read/Write Workloads". 3 August 2022. Retrieved 2022-08-04.
 Turner, J., New directions in communications (or which way to the information age?). IEEE Communications Magazine 24 (10): 8–15. ISSN 0163-6804, 1986.
 Andrew S. Tanenbaum, Computer Networks, Fourth Edition, ISBN 0-13-166836-6, Prentice Hall PTR, 2003., page 401.
 ATM Forum, The User Network Interface (UNI), v. 3.1, ISBN 0-13-393828-X, Prentice Hall PTR, 1995.
 ITU-T, Traffic control and congestion control in B ISDN, Recommendation I.371, International Telecommunication Union, 2004, Annex A, page 87.
 "Linux HTB Home Page". Retrieved 2013-11-30.
Further reading
John Evans, Clarence Filsfils (2007). Deploying IP and MPLS QoS for Multiservice Networks: Theory and Practice. Morgan Kaufmann. ISBN 978-0-12-370549-5.
Ferguson P., Huston G. (1998). Quality of Service: Delivering QoS on the Internet and in Corporate Networks. John Wiley & Sons, Inc. ISBN 0-471-24358-2.
Categories: Network performanceNetwork scheduling algorithms
